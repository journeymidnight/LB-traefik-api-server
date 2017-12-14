package main

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	"golang.org/x/net/context"
	"strings"
	"time"
)

const CONNTIMEOUT = 5 * time.Second
const OPTIMEOUT = 5 * time.Second

func getclient() (error, *clientv3.Client, context.Context) {
	log.Infof("the endponints for etcd is %s", Config.Etcd.Endpoints)
	var cfg clientv3.Config
	if Config.Etcd.Https {
		tlsInfo := transport.TLSInfo{
			CertFile:      Config.Etcd.Certfile,
			KeyFile:       Config.Etcd.Keyfile,
			TrustedCAFile: Config.Etcd.CertCA,
		}
		tlsconfig, _ := tlsInfo.ClientConfig()
		cfg = clientv3.Config{
			Endpoints:   strings.Split(Config.Etcd.Endpoints, ","),
			DialTimeout: CONNTIMEOUT,
			TLS:         tlsconfig,
		}
	} else {
		cfg = clientv3.Config{
			Endpoints:   strings.Split(Config.Etcd.Endpoints, ","),
			DialTimeout: CONNTIMEOUT,
		}
	}
	client, err := clientv3.New(cfg)
	if err != nil {
		log.Error(err)
		return err, nil, nil
	}
	ctx, _ := context.WithTimeout(context.Background(), OPTIMEOUT)
	return nil, client, ctx
}

func List(prefix string) ([]string, error) {
	err, client, ctx := getclient()
	if err != nil {
		return nil, err
	}
	resp, err1 := client.Get(ctx, prefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	if err1 != nil {
		return nil, err1
	}
	result := []string{}
	for _, ev := range resp.Kvs {
		k := strings.TrimPrefix(string(ev.Key), prefix)
		k = strings.TrimPrefix(k, "/")
		tmp := strings.Split(k, "/")[0]
		result = append(result, tmp)
	}
	result = removeRepeat(result)
	return result, nil

}

func Put(key string, value string) error {
	err, client, ctx := getclient()
	if err != nil {
		return err
	}
	defer client.Close()
	_, err = client.Put(ctx, key, value)
	if err != nil {
		return err
	}
	return nil
}

func PutMap(srcmap map[string]string) error {
	err, client, ctx := getclient()
	if err != nil {
		return err
	}
	defer client.Close()
	for k, v := range srcmap {
		if _, err = client.Put(ctx, k, v); err != nil {
			return err
		}
	}
	return nil
}

func Get(key string) (string, error) {
	err, client, ctx := getclient()
	if err != nil {
		return "", err
	}
	defer client.Close()
	resp, err1 := client.Get(ctx, key)
	if err1 != nil {
		return "", err
	}
	if resp.Kvs != nil {
		return string(resp.Kvs[0].Value), nil
	}
	return "", nil
}

func DeleteWithPrefix(prefix string) error {
	err, client, ctx := getclient()
	if err != nil {
		return err
	}
	defer client.Close()
	if _, err1 := client.Delete(ctx, prefix, clientv3.WithPrefix()); err1 != nil {
		return err1
	}
	return nil
}

func DeleteWithPrefixInList(prefixs []string) error {
	err, client, ctx := getclient()
	if err != nil {
		return err
	}
	defer client.Close()
	for _, prefix := range prefixs {
		if _, err := client.Delete(ctx, prefix, clientv3.WithPrefix()); err != nil {
			return err
		}
	}
	return nil
}

func removeRepeat(src []string) []string {
	helpmap := make(map[string]struct{})
	for _, v := range src {
		if _, exsit := helpmap[v]; !exsit {
			helpmap[v] = struct{}{}
		}
	}
	dst := []string{}
	for s := range helpmap {
		dst = append(dst, s)
	}
	return dst
}
