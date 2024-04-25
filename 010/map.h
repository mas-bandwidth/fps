
#include "shared.h"

#define MAP_BUCKET_SIZE                          32
#define MAP_NUM_BUCKETS             PLAYERS_PER_CPU

struct map_element_t
{
    uint64_t session_id;
    uint8_t * player_data;
};

struct map_bucket_t
{
    int size;
    struct map_element_t elements[MAP_BUCKET_SIZE];
};

struct map_t
{
    int size;
    struct map_bucket_t buckets[MAP_NUM_BUCKETS];
};

static void map_reset( struct map_t * map );

struct map_t * map_create()
{
    struct map_t * map = (struct map_t*) malloc( sizeof( struct map_t ) );
    assert( map );
    memset( map, 0, sizeof(map) );
    map_reset( map );
    return map;
}

static void map_destroy( struct map_t * map )
{
    assert( map );
    free( map );
}

static void map_bucket_reset( struct map_bucket_t * bucket )
{
    assert( bucket );
    int i;
    bucket->size = 0;
    for ( i = 0; i < MAP_NUM_BUCKETS; i++ )
    {
        struct map_element_t * element = bucket->elements + i;
        element->session_id = 0;
        if ( element->player_data )
        {
            free( element->player_data );
            element->player_data = NULL;
        }
    }
}

static void map_reset( struct map_t * map )
{
    assert( map );
    int i;
    map->size = 0;
    for ( i = 0; i < MAP_NUM_BUCKETS; i++ )
    {
        struct map_bucket_t * bucket = map->buckets + i;
        netcode_address_map_bucket_reset( bucket );
    }
}

/*
static int netcode_address_map_set( struct netcode_address_map_t * map,
                                    struct netcode_address_t * address,
                                    int client_index )
{
    int bucket_index = netcode_address_hash( address );
    struct netcode_address_map_bucket_t * bucket = map->buckets + bucket_index;
    if ( bucket->size == NETCODE_MAX_CLIENTS )
    {
        return 0;
    }

    struct netcode_address_map_element_t * element = bucket->elements + bucket->size;
    element->client_index = client_index;
    element->address = *address;

    ++bucket->size;
    ++map->size;

    return 1;
}

static struct netcode_address_map_element_t * netcode_address_map_bucket_find(
    struct netcode_address_map_bucket_t * bucket,
    struct netcode_address_t * address )
{
    int i;
    for ( i = 0; i < bucket->size; i++ )
    {
        struct netcode_address_map_element_t * element = bucket->elements + i;
        if ( netcode_address_equal( address, &element->address ) )
        {
            return element;
        }
    }

    return NULL;
}

static int netcode_address_map_get( struct netcode_address_map_t * map,
                                    struct netcode_address_t * address )
{
    int bucket_index = netcode_address_hash( address );
    struct netcode_address_map_bucket_t * bucket = map->buckets + bucket_index;
    struct netcode_address_map_element_t * element = netcode_address_map_bucket_find( bucket, address );
    
    if ( !element )
    {
        return -1;
    }

    return element->client_index;
}

static int netcode_address_map_delete( struct netcode_address_map_t * map,
                                       struct netcode_address_t * address )
{
    int bucket_index = netcode_address_hash( address );
    struct netcode_address_map_bucket_t * bucket = map->buckets + bucket_index;
    struct netcode_address_map_element_t * element = netcode_address_map_bucket_find( bucket, address );

    if ( !element )
    {
        return 0;
    }

    struct netcode_address_map_element_t * last = bucket->elements + (bucket->size - 1);
    *element = *last;
    netcode_address_map_element_reset(last);

    --bucket->size;
    --map->size;

    return 1;
}
*/
