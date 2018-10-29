package controllers

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

const namespace = "vmware"

var (
	prometheusHostPowerState = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "power_state",
		Help:      "poweredOn 1, poweredOff 2, standBy 3, other 0",
	}, []string{"host_name"})
	prometheusHostBoot = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "boot_timestamp_seconds",
		Help:      "Uptime host",
	}, []string{"host_name"})
	prometheusTotalCpu = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "cpu_max",
		Help:      "CPU total",
	}, []string{"host_name"})
	prometheusUsageCpu = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "cpu_usage",
		Help:      "CPU Free",
	}, []string{"host_name"})
	prometheusTotalMem = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "memory_max",
		Help:      "Memory max",
	}, []string{"host_name"})
	prometheusUsageMem = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "memory_usage",
		Help:      "Memory Free",
	}, []string{"host_name"})
	prometheusTotalDs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "datastore",
		Name:      "capacity_size",
		Help:      "Datastore total",
	}, []string{"ds_name"})
	prometheusUsageDs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "datastore",
		Name:      "freespace_size",
		Help:      "Datastore free",
	}, []string{"ds_name"})
	prometheusVmBoot = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "boot_timestamp_seconds",
		Help:      "VMWare VM boot time in seconds",
	}, []string{"vm_name"})
	prometheusVmCpuAval = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "cpu_avaleblemhz",
		Help:      "VMWare VM usage CPU",
	}, []string{"vm_name"})
	prometheusVmCpuUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "cpu_usagemhz",
		Help:      "VMWare VM usage CPU",
	}, []string{"vm_name"})
	prometheusVmNumCpu = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "num_cpu",
		Help:      "Available number of cores",
	}, []string{"vm_name"})
	prometheusVmMemAval = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "mem_avaleble",
		Help:      "Available memory",
	}, []string{"vm_name"})
	prometheusVmMemUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "mem_usage",
		Help:      "Usage memory",
	}, []string{"vm_name"})
	prometheusVmNetRec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "net_rec",
		Help:      "Usage memory",
	}, []string{"vm_name"})
)

func totalCpu(hs mo.HostSystem) float64 {
	totalCPU := int64(hs.Summary.Hardware.CpuMhz) * int64(hs.Summary.Hardware.NumCpuCores)
	return float64(totalCPU)
}

func freeCPU(hs mo.HostSystem) float64 {
	totalCPU := int64(hs.Summary.Hardware.CpuMhz) * int64(hs.Summary.Hardware.NumCpuCores)
	freeCPU := int64(totalCPU) - int64(hs.Summary.QuickStats.OverallCpuUsage)
	return float64(freeCPU)
}

func powerState(s types.HostSystemPowerState) float64 {
	if s == "poweredOn" {
		return 1
	}
	if s == "poweredOff" {
		return 2
	}
	if s == "standBy" {
		return 3
	}
	return 0
}

func RegistredMetrics() {
	prometheus.MustRegister(
		prometheusHostPowerState,
		prometheusHostBoot,
		prometheusTotalCpu,
		prometheusUsageCpu,
		prometheusTotalMem,
		prometheusUsageMem,
		prometheusTotalDs,
		prometheusUsageDs,
		prometheusVmBoot,
		prometheusVmCpuAval,
		prometheusVmNumCpu,
		prometheusVmMemAval,
		prometheusVmMemUsage,
		prometheusVmCpuUsage,
		prometheusVmNetRec)
}

func NewVmwareHostMetrics(host string, username string, password string) {
	url := NewURL(host, username, password)
	ctx := context.Background()
	c, err := NewClient(ctx, url)
	if err != nil {
		log.Fatal(err)
	}

	defer c.Logout(ctx)

	m := view.NewManager(c.Client)

	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		log.Fatal(err)
	}

	defer v.Destroy(ctx)

	var hss []mo.HostSystem
	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hss)
	if err != nil {
		log.Fatal(err)
	}

	for _, hs := range hss {
		prometheusHostPowerState.WithLabelValues(hs.Summary.Config.Name).Set(powerState(hs.Summary.Runtime.PowerState))
		prometheusHostBoot.WithLabelValues(hs.Summary.Config.Name).Set(float64(hs.Summary.Runtime.BootTime.Unix()))
		prometheusTotalCpu.WithLabelValues(hs.Summary.Config.Name).Set(totalCpu(hs))
		prometheusUsageCpu.WithLabelValues(hs.Summary.Config.Name).Set(freeCPU(hs))
		prometheusTotalMem.WithLabelValues(hs.Summary.Config.Name).Set(float64(hs.Summary.Hardware.MemorySize))
		prometheusUsageMem.WithLabelValues(hs.Summary.Config.Name).Set(float64(hs.Summary.QuickStats.OverallMemoryUsage) * 1024 * 1024)
	}

}

func NewVmwareDsMetrics(host string, username string, password string) {
	url := NewURL(host, username, password)
	ctx := context.Background()
	c, err := NewClient(ctx, url)
	if err != nil {
		log.Fatal(err)
	}

	defer c.Logout(ctx)

	m := view.NewManager(c.Client)

	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Datastore"}, true)
	if err != nil {
		log.Fatal(err)
	}

	defer v.Destroy(ctx)

	var dss []mo.Datastore
	err = v.Retrieve(ctx, []string{"Datastore"}, []string{"summary"}, &dss)
	if err != nil {
		log.Fatal(err)
	}

	for _, ds := range dss {
		prometheusTotalDs.WithLabelValues(ds.Summary.Name).Set(float64(ds.Summary.Capacity))
		prometheusUsageDs.WithLabelValues(ds.Summary.Name).Set(float64(ds.Summary.FreeSpace))
	}

}

func NewVmwareVmMetrics(host string, username string, password string) {
	url := NewURL(host, username, password)
	ctx := context.Background()
	c, err := NewClient(ctx, url)
	if err != nil {
		log.Fatal(err)
	}

	defer c.Logout(ctx)

	m := view.NewManager(c.Client)

	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		log.Fatal(err)
	}

	defer v.Destroy(ctx)

	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms)
	if err != nil {
		log.Fatal(err)
	}

	for _, vm := range vms {
		prometheusVmBoot.WithLabelValues(vm.Summary.Config.Name).Set(float64(vm.Summary.Runtime.BootTime.Unix()))
		prometheusVmCpuAval.WithLabelValues(vm.Summary.Config.Name).Set(float64(vm.Summary.Runtime.MaxCpuUsage) * 1000 * 1000)
		prometheusVmCpuUsage.WithLabelValues(vm.Summary.Config.Name).Set(float64(vm.Summary.QuickStats.OverallCpuUsage) * 1000 * 1000)
		prometheusVmNumCpu.WithLabelValues(vm.Summary.Config.Name).Set(float64(vm.Summary.Config.NumCpu))
		prometheusVmMemAval.WithLabelValues(vm.Summary.Config.Name).Set(float64(vm.Summary.Config.MemorySizeMB))
		prometheusVmMemUsage.WithLabelValues(vm.Summary.Config.Name).Set(float64(vm.Summary.QuickStats.GuestMemoryUsage) * 1024 * 1204)
	}

}
