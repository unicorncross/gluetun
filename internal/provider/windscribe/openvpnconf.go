package windscribe

import (
	"strconv"
	"strings"

	"github.com/qdm12/gluetun/internal/configuration"
	"github.com/qdm12/gluetun/internal/constants"
	"github.com/qdm12/gluetun/internal/models"
	"github.com/qdm12/gluetun/internal/provider/utils"
)

func (w *Windscribe) BuildConf(connection models.Connection,
	settings configuration.OpenVPN) (lines []string, err error) {
	if settings.Cipher == "" {
		settings.Cipher = constants.AES256gcm
	}

	if settings.Auth == "" {
		settings.Auth = constants.SHA512
	}

	lines = []string{
		"client",
		"dev " + settings.Interface,
		"nobind",
		"remote-cert-tls server",
		"ping 10",
		"ping-exit 60",
		"tls-exit",

		// Windscribe specific
		"key-direction 1",
		"reneg-sec 0",

		// Added constant values
		"auth-nocache",
		"mute-replay-warnings",
		"pull-filter ignore \"auth-token\"", // prevent auth failed loops
		"auth-retry nointeract",
		"suppress-timestamps",

		// Modified variables
		"verb " + strconv.Itoa(settings.Verbosity),
		"auth-user-pass " + constants.OpenVPNAuthConf,
		connection.OpenVPNProtoLine(),
		connection.OpenVPNRemoteLine(),
		"auth " + settings.Auth,
		"verify-x509-name " + connection.Hostname + " name",
	}

	lines = append(lines, utils.CipherLines(settings.Cipher, settings.Version)...)

	if strings.HasSuffix(settings.Cipher, "-gcm") {
		lines = append(lines, "ncp-ciphers AES-256-GCM:AES-256-CBC:AES-128-GCM")
	}

	if connection.Protocol == constants.UDP {
		lines = append(lines, "explicit-exit-notify")
	}

	if !settings.Root {
		lines = append(lines, "user "+settings.ProcUser)
		lines = append(lines, "persist-tun")
		lines = append(lines, "persist-key")
	}

	if settings.MSSFix > 0 {
		lines = append(lines, "mssfix "+strconv.Itoa(int(settings.MSSFix)))
	}

	if !settings.IPv6 {
		lines = append(lines, `pull-filter ignore "route-ipv6"`)
		lines = append(lines, `pull-filter ignore "ifconfig-ipv6"`)
	}

	lines = append(lines, utils.WrapOpenvpnCA(
		constants.WindscribeCertificate)...)
	lines = append(lines, utils.WrapOpenvpnTLSAuth(
		constants.WindscribeOpenvpnStaticKeyV1)...)

	lines = append(lines, "")

	return lines, nil
}
