// +build windows

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	clr "github.com/ropnop/go-clr"
)

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func checkOK(hr uintptr, caller string) {
	if hr != 0x0 {
		log.Fatalf("%s returned 0x%08x", caller, hr)
	}
}

func init() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: EXEfromMemory.exe <exe_file> <exe_args>")
		os.Exit(1)
	}
}

func main() {
	filename := os.Args[1]
	exebytes, err := ioutil.ReadFile(filename)
	must(err)
	runtime.KeepAlive(exebytes)

	var params []string
	if len(os.Args) > 2 {
		params = os.Args[2:]
	}

	metaHost, err := clr.CLRCreateInstance(clr.CLSID_CLRMetaHost, clr.IID_ICLRMetaHost)
	must(err)

	versionString := "v4.0.30319"
	pwzVersion, err := syscall.UTF16PtrFromString(versionString)
	must(err)
	runtimeInfo, err := metaHost.GetRuntime(pwzVersion, clr.IID_ICLRRuntimeInfo)
	must(err)

	var isLoadable bool
	err = runtimeInfo.IsLoadable(&isLoadable)
	must(err)
	if !isLoadable {
		log.Fatal("[!] IsLoadable returned false. Bailing...")
	}

	err = runtimeInfo.BindAsLegacyV2Runtime()
	must(err)

	var runtimeHost *clr.ICORRuntimeHost
	err = runtimeInfo.GetInterface(clr.CLSID_CorRuntimeHost, clr.IID_ICorRuntimeHost, unsafe.Pointer(&runtimeHost))
	must(err)
	err = runtimeHost.Start()
	must(err)
	fmt.Println("[+] Loaded CLR into this process")

	var pAppDomain uintptr
	//var pIUnknown uintptr
	iu, err := runtimeHost.GetDefaultDomain()
	must(err)

	//iu := clr.NewIUnknownFromPtr(uintptr(unsafe.Pointer(appDomain2)))
	hr := iu.QueryInterface(&clr.IID_AppDomain, &pAppDomain)
	checkOK(hr, "iu.QueryInterface")
	appDomain := clr.NewAppDomainFromPtr(pAppDomain)
	fmt.Println("[+] Got default AppDomain")

	safeArray, err := clr.CreateSafeArray(exebytes)
	must(err)
	runtime.KeepAlive(safeArray)
	fmt.Println("[+] Created SafeArray from byte array")

	assembly, err := appDomain.Load_3(safeArray)
	must(err)
	fmt.Printf("[+] Loaded %d bytes into memory from %s\n", len(exebytes), filename)
	fmt.Printf("[+] Executable loaded into memory at %p\n", assembly)

	var pEntryPointInfo uintptr
	hr = assembly.GetEntryPoint(&pEntryPointInfo)
	checkOK(hr, "assembly.GetEntryPoint")
	fmt.Printf("[+] Executable entrypoint found at 0x%x\n", pEntryPointInfo)
	methodInfo := clr.NewMethodInfoFromPtr(pEntryPointInfo)

	var methodSignaturePtr, paramPtr uintptr
	err = methodInfo.GetString(&methodSignaturePtr)
	if err != nil {
		return
	}
	methodSignature := clr.ReadUnicodeStr(unsafe.Pointer(methodSignaturePtr))
	fmt.Printf("[+] Checking if the assembly requires arguments\n")
	if !strings.Contains(methodSignature, "Void Main()") {
		if len(params) < 1 {
			log.Fatal("the assembly requires arguments but none were provided\nUsage: EXEfromMemory.exe <exe_file> <exe_args>")
		}
		if paramPtr, err = clr.PrepareParameters(params); err != nil {
			log.Fatal(fmt.Sprintf("there was an error preparing the assembly arguments:\r\n%s", err))
		}
	}

	var pRetCode uintptr
	nullVariant := clr.Variant{
		VT:  1,
		Val: uintptr(0),
	}
	fmt.Println("[+] Invoking...")
	hr = methodInfo.Invoke_3(
		nullVariant,
		paramPtr,
		&pRetCode)

	fmt.Println("-------")

	checkOK(hr, "methodInfo.Invoke_3")
	fmt.Printf("[+] Executable returned code %d\n", pRetCode)

	appDomain.Release()
	runtimeHost.Release()
	runtimeInfo.Release()
	metaHost.Release()

}
