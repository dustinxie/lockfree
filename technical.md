
**Clarification:**
The term "lock-free" does not mean the implementation is entirely without lock,
but that the heavy-lifting workloads of adding/modifying an entry in the map is
done without the need to lock the entire map. A read lock is still used to acess
the buckets of map, but this is much more lightweight and efficient than locking
the entire map for every insert/modify operation.

## Motivation
Golang's native map is not designed to be thread-safe at first place, so you'll
need to do concurrency control where multiple go-routines/threads are accessing
the same map.

Golang provides sync.Map as a solution, which uses a `sync.Mutex` to coordinate
access to the map. However, this is a lock at the entire map level and somehow
inefficient.

## How it works
The basic idea is quite straightforward: implement the map as a linked-list that
is sorted by each entry's hash, and use [CAS](https://en.wikipedia.org/wiki/Compare-and-swap)
operation to insert new entries into the list, thus achieving concurrent access
to the map.

In order to efficiently access the map, special nodes (called fence) are inserted
into the list, breaking the list into multiple chunks. Each chunk begins with its
fence node, with its last node linking to next chunk's fence node. These chunks 
constitute **buckets** of our hashmap.

![image](/pictures/map-structure.jpg)

For the sake of simplicity, let's assume the hash is 8-bit (a number in [0, 255])
and the resulting map holds 256 entries in the following explanation. The actual
implementation used 64-bit long hash so the map can store 2^64 entries, which
should suffice any conceivable use-case. 

### map.Get() is quite easy
To get an existing entry from the map, first determine which bucket it belongs
to according to its hash value
```
func (h *hmap) Get(key) {
    hash := h.hash(key)
    bucket := h.getBucket(hash)
    return bucket.get(key, hash)
}
``` 
`bucket.get` simply runs through the bucket until either the entry is found or
the next bucket's fence node is hit.

### map.Set() using CAS
To insert a new entry `<key, value>` to the map, a hash is first computed using
`key` as input to a hash function `hash = h(key)`. The resulting hash value is
then used to find out the insertion position in the bucket.

In our example below, a map keeps track of the number of animals living in a
happy farm. We want to add a new entry that there are 9 rats (`hash = 131`). By
searching the hash value, this new entry is to be inserted between `A` and `B`.

At the same time, a second go-routine/thread is also reporting that there are 97
ants (with `hash = 136`), and its position is also between `A` and `B`, as shown
in the figure below: 

![image](/pictures/map-before-insert.jpg)

Here's where the CAS kicks in -- let the 2 go-routines compete for linking `A`
to itself, whoever wins the CAS instruction gets its entry inserted, while the
other would fail the CAS and attempt again.
```
for {
	A, B = search(node)
	node.next = B
	if CAS(A.next, B, node) {
		// success
		return
	}
}
``` 
Assume go-routine 1 wins the race, the bucket now looks like:

![image](/pictures/map-afert-insert.jpg)

Note the 2 red arrows indicating change in node linkage. Now `A.next` is equal
to the newly inserted node (`hash = 131`), instead of `B`. Hence go-routine 2
would fail the CAS on its first attempt and go around the for loop again, and
this time it will successfully insert its entry into the bucket.

### map.Del() is a little tricky
Deleting an entry requires changing `prev` to link to `next` atomically (with
`curr` being the node to be deleted), this cannot be safely done with only CAS
because at anytime, new node could be inserted between `prev` and `curr`, or
between `curr` and `next`. 

![image](/pictures/map-delete.jpg)

This actually comes down to a double-CAS operation if we want to do it in one-
shot, something like:
```
func doubleCAS(prev, curr, next) bool {
	// do the following atomically
	if prev.next == curr && curr.next == next {
		prev.next = next
		return true
	}
	return false
}
``` 
I haven't seen such implementation in Golang. Please let me know if you know of
a good one. Hence the current implementation is to lock the entire map for the
delete operation.

An alternative solution is to mark the node as tombstone, and deleting them
during the buckets grow/shrink stage.

However, benchmark result shows that this solution is even slightly slower. This
is mostly due to the following 2 reasons:
1. `Get()` and `Set()` has to add an additional check if an entry is marked as
tombstone
2.  the cost of checking and removing tombstone during bucket grow/shrink

Looks like the benefit gained by not locking the map cannot compensate for the 
penalty incurred.

## No re-hash, just grow/shrink the buckets
As more and more entries are added to the map, the time it takes to search an
entry in the bucket keeps growing. Once the average size of bucket exceeds a
threshold value, another fence node is added in the middle.

The figure below shows inserting a new fence node (`hash = 128`). Now access to
all entries with `hash >= 128` would start from this new node. This effectively
splits the bucket into 2, and reduces the search/insert time by half.

![image](/pictures/map-split.jpg)

And next time as the 2 buckets become crowded enough, another split is triggered
to insert 2 more fence nodes (`hash = 64 and 192`), making a total of 4 buckets.

In the same fashion, when enough entries are deleted from the map, the average
size of bucket keeps decreasing and once it drops below a threshold, we don't
need that many buckets. When that happens, 2 consecutive buckets are merged back
into one by removing the fence node in the middle -- an exact opposite to the
split operation.

This split/merge is very fast since it's no more than adding/removing particular
nodes to/from the list. This approach eradicates the costly data copy between
buckets usually seen in the re-hash stage of a traditional hash table, turning
re-hash to very simple and fast bucket adjustment as the toal number of entries
in the map varies along time.

## Hash-flooding attack
If an attacker manages to create many collision keys or keys that all hash to a
limited range (for instance 0~65535), that is a so-called hash-flooding DoS
attack and can severely degrade the hashmap's performance.

Our implementation used 64-bit [SipHash](https://en.wikipedia.org/wiki/SipHash)
to compute hash from key. The SipHash is immune to this attack, and also more 
efficient than a cryptographically secure hash function (such as SHA256).
