// mysqltrace_kern.c
#include "vmlinux.h"
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

#define ETH_P_IP 0x0800
#define IPPROTO_TCP 6
#define MYSQL_PORT 3306
#define CAPTURE_BYTES 64

struct event {
  __u32 src_ip;
  __u32 dst_ip;
  __u16 src_port;
  __u16 dst_port;
  __u32 len;
  unsigned char payload[CAPTURE_BYTES];
};

struct {
  __uint(type, BPF_MAP_TYPE_RINGBUF);
  __uint(max_entries, 1 << 24);
} events SEC(".maps");

SEC("xdp")
int trace_mysql_response(struct xdp_md *ctx) {
  void *data_end = (void *)(long)ctx->data_end;
  void *data = (void *)(long)ctx->data;
  struct ethhdr *eth = data;

  if ((void *)(eth + 1) > data_end) {
    // bpf_printk("Packet too short for ethhdr\n");
    return XDP_PASS;
  }

  if (bpf_ntohs(eth->h_proto) != ETH_P_IP) {
    // bpf_printk("Not an IP packet\n");
    return XDP_PASS;
  }

  struct iphdr *ip = (void *)(eth + 1);
  if ((void *)(ip + 1) > data_end) {
    // bpf_printk("Packet too short for iphdr\n");
    return XDP_PASS;
  }

  if (ip->protocol != IPPROTO_TCP) {
    // bpf_printk("Not a TCP packet\n");
    return XDP_PASS;
  }

  int ip_header_len = ip->ihl * 4;
  struct tcphdr *tcp = (void *)ip + ip_header_len;
  if ((void *)(tcp + 1) > data_end) {
    // bpf_printk("Packet too short for tcphdr\n");
    return XDP_PASS;
  }

  __u16 sport = bpf_ntohs(tcp->source);
  __u16 dport = bpf_ntohs(tcp->dest);
  __u32 saddr = bpf_ntohl(ip->saddr);
  __u32 daddr = bpf_ntohl(ip->daddr);

  // detect MySQL response (src port 3306)
  if (dport != MYSQL_PORT) {
    bpf_printk("Not a MySQL response packet\n");
    return XDP_PASS;
  }
  bpf_printk("TCP %pI4:%d ->%pI4:%d\n", &ip->saddr, sport, &ip->daddr, dport);

  // bpf_printk("Packet by port 3306 received\n");

  struct event *evt = bpf_ringbuf_reserve(&events, sizeof(*evt), 0);
  if (!evt)
    return XDP_PASS;

  evt->src_ip = saddr;
  evt->dst_ip = daddr;
  evt->src_port = sport;
  evt->dst_port = dport;

  void *payload = (void *)(tcp + 1);
  if (payload < data_end) {
    __u32 to_read = data_end - payload;
    if (to_read > CAPTURE_BYTES)
      to_read = CAPTURE_BYTES;
    // bpf_probe_read_kernel(evt->payload, to_read, payload);
    evt->len = to_read;
  } else {
    evt->len = 0;
  }

  bpf_ringbuf_submit(evt, 0);
  return XDP_PASS;
}

char LICENSE[] SEC("license") = "GPL";
