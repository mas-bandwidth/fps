
#include <linux/in.h>
#include <linux/if_ether.h>
#include <linux/if_packet.h>
#include <linux/if_vlan.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/udp.h>
#include <linux/bpf.h>
#include <linux/string.h>
#include <bpf/bpf_helpers.h>

#if defined(__BYTE_ORDER__) && defined(__ORDER_LITTLE_ENDIAN__) && \
    __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
#define bpf_ntohs(x)        __builtin_bswap16(x)
#define bpf_htons(x)        __builtin_bswap16(x)
#elif defined(__BYTE_ORDER__) && defined(__ORDER_BIG_ENDIAN__) && \
    __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
#define bpf_ntohs(x)        (x)
#define bpf_htons(x)        (x)
#else
# error "Endianness detection needs to be set up for your compiler?!"
#endif

struct inner_player_state_map {
    __uint( type, BPF_MAP_TYPE_LRU_HASH );
    __type( key, __u64 );
    __type( value, struct player_state );
    __uint( max_entries, MAX_SESSIONS / MAX_CPUS );
} 
player_state_0 SEC(".maps"),
player_state_1 SEC(".maps"),
player_state_2 SEC(".maps"),
player_state_3 SEC(".maps"),
player_state_4 SEC(".maps"),
player_state_5 SEC(".maps"),
player_state_6 SEC(".maps"),
player_state_7 SEC(".maps"),
player_state_8 SEC(".maps"),
player_state_9 SEC(".maps"),
player_state_10 SEC(".maps"),
player_state_11 SEC(".maps"),
player_state_12 SEC(".maps"),
player_state_13 SEC(".maps"),
player_state_14 SEC(".maps"),
player_state_15 SEC(".maps");

struct {
    __uint( type, BPF_MAP_TYPE_ARRAY_OF_MAPS );
    __uint( max_entries, MAX_CPUS );
    __type( key, __u32 );
    __uint( pinning, LIBBPF_PIN_BY_NAME );
    __array( values, struct inner_player_state_map );
} player_state_map SEC(".maps") = {
    .values = { 
        &player_state_0,
        &player_state_1,
        &player_state_2,
        &player_state_3,
        &player_state_4,
        &player_state_5,
        &player_state_6,
        &player_state_7,
        &player_state_8,
        &player_state_9,
        &player_state_10,
        &player_state_11,
        &player_state_12,
        &player_state_13,
        &player_state_14,
        &player_state_15,
    }
};

SEC("server_xdp") int server_xdp_filter( struct xdp_md *ctx ) 
{ 
    void * data = (void*) (long) ctx->data; 

    void * data_end = (void*) (long) ctx->data_end; 

    struct ethhdr * eth = data;

    if ( (void*)eth + sizeof(struct ethhdr) < data_end )
    {
        if ( eth->h_proto == __constant_htons(ETH_P_IP) ) // IPV4
        {
            struct iphdr * ip = data + sizeof(struct ethhdr);

            if ( (void*)ip + sizeof(struct iphdr) < data_end )
            {
                if ( ip->protocol == IPPROTO_UDP ) // UDP
                {
                    struct udphdr * udp = (void*) ip + sizeof(struct iphdr);

                    if ( (void*)udp + sizeof(struct udphdr) <= data_end )
                    {
                        if ( udp->dest == __constant_htons(40000) )
                        {
                            return XDP_DROP;
                        }
                    }
                }
            }
        }
    }

    return XDP_PASS;
}

char _license[] SEC("license") = "GPL";
