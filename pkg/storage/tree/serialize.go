package tree

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"io"

	"github.com/pyroscope-io/pyroscope/pkg/storage/dict"
	"github.com/pyroscope-io/pyroscope/pkg/util/varint"
)

// serialization format version. it's not very useful right now, but it will be in the future
const currentVersion = 1

func (t *Tree) Serialize(d *dict.Dict, maxNodes int, w io.Writer) error {
	t.Lock()
	defer t.Unlock()
	vw := varint.NewWriter()

	// TODO: introduce the next version, ensure backward compatibility.
	// TODO: implement truncation (?).
	var err error
	if _, err = vw.Write(w, currentVersion); err != nil {
		return err
	}

	if _, err = vw.Write(w, uint64(len(t.nodes))); err != nil {
		return err
	}

	for _, n := range t.nodes {
		if _, err = vw.Write(w, uint64(len(n.ChildrenNodes))); err != nil {
			return err
		}
		for _, c := range n.ChildrenNodes {
			if _, err = vw.Write(w, uint64(c)); err != nil {
				return err
			}
		}
		// TODO: labels to be written as a blob.
		// TODO: what if not use dictionaries?
		labelKey := d.Put(t.loadLabel(n.labelPosition))
		if _, err = vw.Write(w, uint64(len(labelKey))); err != nil {
			return err
		}
		if _, err = w.Write(labelKey); err != nil {
			return err
		}
		if _, err = vw.Write(w, n.Self); err != nil {
			return err
		}
		if _, err = vw.Write(w, n.Total); err != nil {
			return err
		}
	}

	return nil
}

func Deserialize(d *dict.Dict, r io.Reader) (*Tree, error) {
	br := bufio.NewReader(r)

	// reads serialization format version, see comment at the top
	var err error
	if _, err = binary.ReadUvarint(br); err != nil {
		return nil, err
	}

	var nodes uint64
	if nodes, err = varint.Read(br); err != nil {
		return nil, err
	}

	t := New()
	if s := int(nodes) - cap(t.nodes); s > 0 {
		t.grow(s)
	}

	var nameBuf bytes.Buffer
	for i := uint64(0); i < nodes; i++ {
		var cl uint64
		if cl, err = binary.ReadUvarint(br); err != nil {
			return nil, err
		}
		var n treeNode
		for j := uint64(0); j < cl; j++ {
			var c uint64
			if c, err = binary.ReadUvarint(br); err != nil {
				return nil, err
			}
			n.ChildrenNodes = append(n.ChildrenNodes, int(c))
		}

		var labelLen uint64
		if labelLen, err = varint.Read(br); err != nil {
			return nil, err
		}

		labelLinkBuf := make([]byte, labelLen)
		if _, err = io.ReadAtLeast(br, labelLinkBuf, int(labelLen)); err != nil {
			return nil, err
		}

		nameBuf.Reset()
		if !d.GetValue(labelLinkBuf, &nameBuf) {
			// these strings has to be at least slightly different, hence base64 Addon
			nameBuf.Reset()
			nameBuf.WriteString("label not found " + base64.URLEncoding.EncodeToString(labelLinkBuf))
		}
		n.labelPosition = t.insertLabel(nameBuf.Bytes())

		if n.Self, err = binary.ReadUvarint(br); err != nil {
			return nil, err
		}
		if n.Total, err = binary.ReadUvarint(br); err != nil {
			return nil, err
		}
		t.nodes = append(t.nodes, n)
	}

	return t, nil
}
