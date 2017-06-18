package tree

import (
	"reflect"
	"testing"
)

func TestNewNode(t *testing.T) {
	type args struct {
		key    string
		parent *Node
	}

	parent := NewNode("parent", nil)

	tests := []struct {
		name string
		args args
		want *Node
	}{
		{
			name: "Creates correct node",
			args: args{
				key:    "test",
				parent: parent,
			},
			want: &Node{
				Key:      "test",
				Children: nodeChildren{},
				Parent:   parent,
				Value:    nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNode(tt.args.key, tt.args.parent); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNode_buildPathUpwards(t *testing.T) {
	type fields struct {
		Key      string
		Value    []byte
		Parent   *Node
		Children nodeChildren
	}

	parent := NewNode(TREE_ROOT, nil)

	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "Checks if building upward path correctly",
			fields: fields{
				Key:      "test",
				Value:    nil,
				Parent:   parent,
				Children: nil,
			},
			want: []string{"test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Node{
				Key:      tt.fields.Key,
				Value:    tt.fields.Value,
				Parent:   tt.fields.Parent,
				Children: tt.fields.Children,
			}

			if got := n.buildPathUpwards(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Node.buildPathUpwards() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNode_Insert(t *testing.T) {
	type fields struct {
		Key      string
		Value    []byte
		Parent   *Node
		Children nodeChildren
	}
	type args struct {
		keys []string
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    OpInfo
		wantErr bool
	}{
		{
			name: "Is Inserting correct",
			fields: fields{
				Key:      TREE_ROOT,
				Value:    nil,
				Parent:   nil,
				Children: nodeChildren{},
			},
			args: args{
				keys: []string{"foo", "bar"},
				data: []byte("foobar"),
			},
			want: OpInfo{
				Key:   "bar",
				Path:  []string{"foo", "bar"},
				Value: []byte("foobar"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Node{
				Key:      tt.fields.Key,
				Value:    tt.fields.Value,
				Parent:   tt.fields.Parent,
				Children: tt.fields.Children,
			}
			got, err := n.Insert(tt.args.keys, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Node.Insert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Node.Insert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNode_Get(t *testing.T) {
	type fields struct {
		Key      string
		Value    []byte
		Parent   *Node
		Children map[string]*Node
	}
	type args struct {
		path []string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Gets correct",
			fields: fields{
				Key:      TREE_ROOT,
				Value:    nil,
				Parent:   nil,
				Children: nodeChildren{},
			},
			args: args{
				path: []string{"foo", "bar"},
			},
			want:    []byte("foobar"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := Node{
				Key:      tt.fields.Key,
				Value:    tt.fields.Value,
				Parent:   tt.fields.Parent,
				Children: tt.fields.Children,
			}

			n.Insert([]string{"foo", "bar"}, []byte("foobar"))
			got, err := n.Get(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Node.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Node.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}
