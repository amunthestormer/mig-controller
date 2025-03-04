package directvolumemigration

import (
	"fmt"
	"net/url"

	"github.com/konveyor/mig-controller/pkg/settings"

	//"encoding/asn1"

	//"k8s.io/apimachinery/pkg/types"

	cranemeta "github.com/konveyor/crane-lib/state_transfer/meta"
	cranetransport "github.com/konveyor/crane-lib/state_transfer/transport"
	stunneltransport "github.com/konveyor/crane-lib/state_transfer/transport/stunnel"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type stunnelConfig struct {
	Name          string
	Namespace     string
	StunnelPort   int32
	RsyncRoute    string
	RsyncPort     int32
	VerifyCA      bool
	VerifyCALevel string
	stunnelProxyConfig
}

type stunnelProxyConfig struct {
	ProxyHost     string
	ProxyUsername string
	ProxyPassword string
}

func (t *Task) ensureStunnelTransport() error {
	destClient, err := t.getDestinationClient()
	if err != nil {
		return err
	}

	// Get client for source
	srcClient, err := t.getSourceClient()
	if err != nil {
		return err
	}

	transportOptions, err := t.getStunnelOptions()
	if err != nil {
		return err
	}

	for ns := range t.getPVCNamespaceMap() {
		sourceNs := getSourceNs(ns)
		destNs := getDestNs(ns)
		nnPair := cranemeta.NewNamespacedPair(
			types.NamespacedName{Name: DirectVolumeMigrationRsyncClient, Namespace: sourceNs},
			types.NamespacedName{Name: DirectVolumeMigrationRsyncClient, Namespace: destNs},
		)

		endpoint, _ := t.getEndpoint(destClient, destNs)
		if endpoint == nil {
			continue
		}

		stunnelTransport, err := stunneltransport.GetTransportFromKubeObjects(
			srcClient, destClient, nnPair, endpoint, transportOptions)
		if err != nil && !k8serror.IsNotFound(err) {
			return err
		}

		if stunnelTransport == nil {
			nsPair := cranemeta.NewNamespacedPair(
				types.NamespacedName{Namespace: sourceNs},
				types.NamespacedName{Namespace: destNs},
			)
			stunnelTransport = stunneltransport.NewTransport(nsPair, transportOptions)

			err = stunnelTransport.CreateServer(destClient, endpoint)
			if err != nil {
				return err
			}

			err = stunnelTransport.CreateClient(srcClient, endpoint)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *Task) getStunnelOptions() (*cranetransport.Options, error) {
	proxyConfig, err := t.generateStunnelProxyConfig()
	if err != nil {
		return nil, err
	}
	transportOptions := &cranetransport.Options{
		ProxyURL:      proxyConfig.ProxyHost,
		ProxyUsername: proxyConfig.ProxyUsername,
		ProxyPassword: proxyConfig.ProxyPassword,
	}
	// retrieve transfer image from source cluster
	srcCluster, err := t.Owner.GetSourceCluster(t.Client)
	if err != nil {
		return nil, err
	}
	if srcCluster != nil {
		srcTransferImage, err := srcCluster.GetRsyncTransferImage(t.Client)
		if err != nil {
			return nil, err
		}
		transportOptions.StunnelClientImage = srcTransferImage
	}
	// retrieve transfer image from destination cluster
	destCluster, err := t.Owner.GetDestinationCluster(t.Client)
	if err != nil {
		return nil, err
	}
	if destCluster != nil {
		destTransferImage, err := destCluster.GetRsyncTransferImage(t.Client)
		if err != nil {
			return nil, err
		}
		transportOptions.StunnelServerImage = destTransferImage
	}
	return transportOptions, nil
}

// generateStunnelProxyConfig loads stunnel proxy configuration from app settings
func (t *Task) generateStunnelProxyConfig() (stunnelProxyConfig, error) {
	var proxyConfig stunnelProxyConfig
	tcpProxyString := settings.Settings.DvmOpts.StunnelTCPProxy
	if tcpProxyString != "" {
		t.Log.Info("Found TCP proxy string. Configuring Stunnel proxy.",
			"tcpProxyString", tcpProxyString)
		url, err := url.Parse(tcpProxyString)
		if err != nil {
			t.Log.Error(err, fmt.Sprintf("failed to parse %s setting", settings.TCPProxyKey))
			return proxyConfig, err
		}
		proxyConfig.ProxyHost = url.Host
		if url.User != nil {
			proxyConfig.ProxyUsername = url.User.Username()
			if pass, set := url.User.Password(); set {
				proxyConfig.ProxyPassword = pass
			}
		}
	}
	return proxyConfig, nil
}
