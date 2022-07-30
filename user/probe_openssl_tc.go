package user

import (
	"errors"
	"github.com/cilium/ebpf"
	manager "github.com/ehids/ebpfmanager"
	"golang.org/x/sys/unix"
	"math"
)

func (this *MOpenSSLProbe) setupManagersTC() error {
	var ifname, binaryPath string

	ifname = this.conf.(*OpensslConfig).Ifname

	switch this.conf.(*OpensslConfig).elfType {
	case ELF_TYPE_BIN:
		binaryPath = this.conf.(*OpensslConfig).Curlpath
	case ELF_TYPE_SO:
		binaryPath = this.conf.(*OpensslConfig).Openssl
	default:
		//如果没找到
		binaryPath = "/lib/x86_64-linux-gnu/libssl.so.1.1"
	}

	this.logger.Printf("%s\tInterface:%s, Pcapng filepath:%s\n", this.Name(), ifname, this.pcapngFilename)

	this.bpfManager = &manager.Manager{
		Probes: []*manager.Probe{
			// customize deleteed TC filter
			// tc filter del dev eth0 ingress
			// tc filter del dev eth0 egress
			{
				Section:          "classifier/egress",
				EbpfFuncName:     "egress_cls_func",
				Ifname:           ifname,
				NetworkDirection: manager.Egress,
			},
			{
				Section:          "classifier/ingress",
				EbpfFuncName:     "ingress_cls_func",
				Ifname:           ifname,
				NetworkDirection: manager.Ingress,
			},
			// --------------------------------------------------

			// openssl masterkey
			{
				Section:          "uprobe/SSL_write_key",
				EbpfFuncName:     "probe_ssl_master_key",
				AttachToFuncName: "SSL_write",
				BinaryPath:       binaryPath,
				UID:              "uprobe_ssl_master_key",
			},
		},

		Maps: []*manager.Map{
			{
				Name: "mastersecret_events",
			},
			{
				Name: "skb_events",
			},
		},
	}

	this.bpfManagerOptions = manager.Options{
		DefaultKProbeMaxActive: 512,

		VerifierOptions: ebpf.CollectionOptions{
			Programs: ebpf.ProgramOptions{
				LogSize: 2097152,
			},
		},

		RLimit: &unix.Rlimit{
			Cur: math.MaxUint64,
			Max: math.MaxUint64,
		},
	}

	if this.conf.EnableGlobalVar() {
		// 填充 RewriteContants 对应map
		this.bpfManagerOptions.ConstantEditors = this.constantEditor()
	}
	return nil
}

func (this *MOpenSSLProbe) initDecodeFunTC() error {
	//SSLDumpEventsMap 与解码函数映射
	SkbEventsMap, found, err := this.bpfManager.GetMap("skb_events")
	if err != nil {
		return err
	}
	if !found {
		return errors.New("cant found map:skb_events")
	}
	this.eventMaps = append(this.eventMaps, SkbEventsMap)
	sslEvent := &TcSkbEvent{}
	sslEvent.SetModule(this)
	this.eventFuncMaps[SkbEventsMap] = sslEvent

	MasterkeyEventsMap, found, err := this.bpfManager.GetMap("mastersecret_events")
	if err != nil {
		return err
	}
	if !found {
		return errors.New("cant found map:mastersecret_events")
	}
	this.eventMaps = append(this.eventMaps, MasterkeyEventsMap)
	masterkeyEvent := &MasterSecretEvent{}
	masterkeyEvent.SetModule(this)
	this.eventFuncMaps[MasterkeyEventsMap] = masterkeyEvent
	return nil
}

func (this *MOpenSSLProbe) dumpTcSkb(event *TcSkbEvent) {
	this.logger.Printf("%s\t%s, length:%d\n", this.Name(), event.String(), event.DataLen)
	return
}
