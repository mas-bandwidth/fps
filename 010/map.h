
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

/*
static void map_reset( struct netcode_address_map_t * map );

struct netcode_address_map_t * netcode_address_map_create( void * allocator_context, 
                                                           void * (*allocate_function)(void*,size_t), 
                                                           void (*free_function)(void*,void*) )
{
    if ( allocate_function == NULL )
    {
        allocate_function = netcode_default_allocate_function;
    }

    if ( free_function == NULL )
    {
        free_function = netcode_default_free_function;
    }

    struct netcode_address_map_t * map = (struct netcode_address_map_t*) 
        allocate_function( allocator_context, sizeof( struct netcode_address_map_t ) );

    netcode_assert( map );

    netcode_address_map_reset( map );

    map->allocator_context = allocator_context;
    map->allocate_function = allocate_function;
    map->free_function = free_function;

    return map;
}

void netcode_address_map_destroy( struct netcode_address_map_t * map )
{
    netcode_assert( map );
    netcode_assert( map->free_function );
    map->free_function( map->allocator_context, map );
}

static void netcode_address_map_element_reset( struct netcode_address_map_element_t * element )
{
    element->client_index = -1;
    memset( &element->address, 0, sizeof( element->address ) );
}

static void map_bucket_reset( struct netcode_address_map_bucket_t * bucket )
{
    int i;
    bucket->size = 0;
    for ( i = 0; i < NETCODE_MAX_CLIENTS; i++ )
    {
        struct netcode_address_map_element_t * element = bucket->elements + i;
        netcode_address_map_element_reset( element );
    }
}

static void netcode_address_map_reset( struct netcode_address_map_t * map )
{
    int i;
    map->size = 0;
    for ( i = 0; i < NETCODE_ADDRESS_MAP_BUCKETS; i++ )
    {
        struct netcode_address_map_bucket_t * bucket = map->buckets + i;
        netcode_address_map_bucket_reset(bucket);
    }
}

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
