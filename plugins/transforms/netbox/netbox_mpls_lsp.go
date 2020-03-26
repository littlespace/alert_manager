package netbox

import (
	"fmt"
	"regexp"

	"github.com/mayuresh82/alert_manager/internal/models"
)

func MplsLspLabels(n *Netbox, alert *models.Alert) (models.Labels, error) {
	labels, err := deviceLabels(n, alert.Device.String)
	if err != nil {
		return nil, err
	}
	// entity(lsp) format = dev1-site1-dev2-site2-num
	pattern := `\w+\d+-\w+\d+-(\w+\d+-\w+\d+)-\d+`
	regEx := regexp.MustCompile(pattern)
	matches := regEx.FindStringSubmatch(alert.Entity)
	if len(matches) != 2 {
		return nil, fmt.Errorf("Could not find a match in lsp name")
	}
	labels["remoteDeviceName"] = matches[1]
	remoteDevLabels, err := deviceLabels(n, matches[1])
	if err != nil {
		return nil, err
	}
	labels["remoteDeviceStatus"] = remoteDevLabels["status"]
	labels["remoteDeviceSite"] = remoteDevLabels["site"]
	return labels, nil
}
