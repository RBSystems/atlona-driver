package atlona

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/byuoitav/common/log"
)

// Amp60 represents an Atlona 60 watt amplifier
type Amp60 struct {
	Address string
}

// AmpStatus represents the current amp status
type AmpStatus struct {
	Model         string `json:"101"`
	Firmware      string `json:"102"`
	MACAddress    string `json:"103"`
	SerialNumber  string `json:"104"`
	OperatingTime string `json:"105"`
}

// AmpAudio represents an audio response from an Atlona 60 watt amp
type AmpAudio struct {
	Volume string `json:"608,omitempty"`
	Muted  string `json:"609,omitempty"`
}

func getR() string {
	return fmt.Sprintf("%v", rand.Float32())
}

func getURL(address, endpoint string) string {
	return "http://" + address + "/action=" + endpoint + "&r=" + getR()
}

func (a *Amp60) sendReq(ctx context.Context, endpoint string) ([]byte, error) {
	var toReturn []byte
	ampUrl := getURL(a.Address, endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", ampUrl, nil)
	if err != nil {
		return toReturn, fmt.Errorf("unable to make new http request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if nerr, ok := err.(*url.Error); ok {
			fmt.Printf("%v\n", nerr.Err)
			if !strings.Contains(nerr.Err.Error(), "malformed") {
				return toReturn, fmt.Errorf("unable to perform request: %w", err)
			}
		} else {
			return toReturn, fmt.Errorf("unable to perform request: %w", err)
		}
		return toReturn, nil
	}
	defer resp.Body.Close()
	toReturn, err = ioutil.ReadAll(resp.Body)
	log.L.Infof("Repsonse: %v\n", resp)

	if err != nil {
		return toReturn, fmt.Errorf("unable to read resp body: %w", err)
	}
	return toReturn, nil
}

// GetInfo gets the current amp status
func (a *Amp60) GetInfo(ctx context.Context) (interface{}, error) {
	resp, err := a.sendReq(ctx, "devicestatus_get")
	if err != nil {
		return nil, fmt.Errorf("unable to get info: %w", err)
	}
	var info AmpStatus
	err = json.Unmarshal(resp, &info)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal into AmpStatus: %w", err)
	}
	return info, nil
}

// GetVolumes gets the current volume
func (a *Amp60) GetVolumes(ctx context.Context, blocks []string) (map[string]int, error) {
	toReturn := make(map[string]int)
	for _, block := range blocks {
		resp, err := a.sendReq(ctx, "deviceaudio_get")
		if err != nil {
			return toReturn, fmt.Errorf("unable to get volume: %w", err)
		}

		var info AmpAudio
		err = json.Unmarshal(resp, &info)
		if err != nil {
			return toReturn, fmt.Errorf("unable to unmarshal into AmpVolume in GetVolume: %w", err)
		}

		volume, err := strconv.Atoi(info.Volume)
		if err != nil {
			return toReturn, fmt.Errorf("error converting volume to int: %s", err)
		}

		toReturn[block] = volume
	}

	return toReturn, nil
}

// GetMutes gets the current muted status
func (a *Amp60) GetMutes(ctx context.Context, blocks []string) (map[string]bool, error) {
	toReturn := make(map[string]bool)

	for _, block := range blocks {
		resp, err := a.sendReq(ctx, "deviceaudio_get")
		if err != nil {
			return toReturn, fmt.Errorf("unable to get muted: %w", err)
		}
		var info AmpAudio
		err = json.Unmarshal(resp, &info)
		if err != nil {
			return toReturn, fmt.Errorf("unable to unmarshal into AmpVolume in GetMuted: %w", err)
		}
		if info.Muted == "1" {
			toReturn[block] = true
		} else {
			toReturn[block] = false
		}
	}

	return toReturn, nil
}

// SetVolume sets the volume on the amp
func (a *Amp60) SetVolume(ctx context.Context, block string, volume int) error {
	_, err := a.sendReq(ctx, fmt.Sprintf("deviceaudio_set&608=%v", volume))
	if err != nil {
		return fmt.Errorf("unable to set volume: %w", err)
	}
	return nil
}

// SetMute sets the current muted status on the amp
func (a *Amp60) SetMute(ctx context.Context, block string, muted bool) error {
	// open a connection with the dsp, set the muted status on block...
	mutedString := "0"
	if muted {
		mutedString = "1"
	}
	_, err := a.sendReq(ctx, fmt.Sprintf("deviceaudio_set&609=%v", mutedString))
	if err != nil {
		return fmt.Errorf("unable to set muted: %w", err)
	}
	return nil
}
