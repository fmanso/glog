package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDocument(t *testing.T) {
	doc := NewDocument("test document", ToDateTime(time.Now().UTC()))
	assert.NotNil(t, doc)
	assert.Equal(t, "test document", doc.Title)
	assert.Equal(t, 0, len(doc.Body))
}

func TestInsertParagraph(t *testing.T) {
	doc := NewDocument("test document", ToDateTime(time.Now().UTC()))

	para1 := doc.InsertParagraph(0, "First paragraph")
	assert.Equal(t, 1, len(doc.Body))
	assert.Equal(t, "First paragraph", string(doc.Body[0].Content))
	assert.Equal(t, para1.ID, doc.Body[0].ID)

	para2 := doc.InsertParagraph(1, "Second paragraph")
	assert.Equal(t, 2, len(doc.Body))
	assert.Equal(t, "Second paragraph", string(doc.Body[1].Content))
	assert.Equal(t, para2.ID, doc.Body[1].ID)

	para3 := doc.InsertParagraph(1, "Inserted paragraph")
	assert.Equal(t, 3, len(doc.Body))
	assert.Equal(t, "Inserted paragraph", string(doc.Body[1].Content))
	assert.Equal(t, para3.ID, doc.Body[1].ID)
}

func TestGetChildrenRoot(t *testing.T) {
	doc := NewDocument("test document", ToDateTime(time.Now().UTC()))
	para1 := doc.InsertParagraph(0, "First paragraph")
	para1.AddChild(NewParagraph("Child paragraph 1"))
	para1.AddChild(NewParagraph("Child paragraph 2"))

	children, found := doc.GetChildren(para1.ID)
	assert.True(t, found)
	assert.Equal(t, 2, len(children)) // para1, para2, and inserted para
	assert.Equal(t, Content("Child paragraph 1"), children[0].Content)
	assert.Equal(t, Content("Child paragraph 2"), children[1].Content)
}

func TestGetChildren(t *testing.T) {
	doc := NewDocument("test document", ToDateTime(time.Now().UTC()))
	para1 := doc.InsertParagraph(0, "First paragraph")
	child1 := NewParagraph("Child paragraph 1")
	para1.AddChild(child1)
	child1.AddChild(NewParagraph("Child paragraph 2"))

	children, found := doc.GetChildren(child1.ID)
	assert.True(t, found)
	assert.Equal(t, 1, len(children)) // para1, para2, and inserted para
	assert.Equal(t, Content("Child paragraph 2"), children[0].Content)
}

func TestGetParentRoot(t *testing.T) {
	doc := NewDocument("test document", ToDateTime(time.Now().UTC()))
	para1 := doc.InsertParagraph(0, "First paragraph")

	parent := doc.GetParentOf(para1.ID)
	assert.Nil(t, parent)
}

func TestGetParent(t *testing.T) {
	doc := NewDocument("test document", ToDateTime(time.Now().UTC()))
	para1 := doc.InsertParagraph(0, "First paragraph")
	child1 := NewParagraph("Child paragraph 1")
	para1.AddChild(child1)
	child2 := NewParagraph("Child paragraph 2")
	child1.AddChild(child2)

	parent := doc.GetParentOf(child2.ID)
	if parent == nil {
		t.Fatal("Expected parent paragraph, got nil")
	}

	assert.Equal(t, child1.ID, parent.ID)
}

func TestIndentRoot(t *testing.T) {
	doc := NewDocument("test document", ToDateTime(time.Now().UTC()))
	para1 := doc.InsertParagraph(0, "First paragraph")
	para2 := doc.InsertParagraph(1, "Second paragraph")

	para, parent, err := doc.Indent(para2.ID)
	assert.NoError(t, err)
	assert.Equal(t, para2.ID, para.ID)
	assert.Equal(t, para1.ID, parent.ID)
	assert.Equal(t, 1, len(para1.Children))
	assert.Equal(t, para2.ID, para1.Children[0].ID)
	assert.Equal(t, 1, len(doc.Body))
}

func TestIndent(t *testing.T) {
	doc := NewDocument("test document", ToDateTime(time.Now().UTC()))
	para1 := doc.InsertParagraph(0, "First paragraph")
	child1 := NewParagraph("Child paragraph 1")
	para1.AddChild(child1)
	child2 := NewParagraph("Child paragraph 2")
	para1.AddChild(child2)

	para, parent, err := doc.Indent(child2.ID)
	assert.NoError(t, err)
	assert.Equal(t, child2.ID, para.ID)
	assert.Equal(t, child1.ID, parent.ID)
	assert.Equal(t, 1, len(para1.Children))
	assert.Equal(t, 1, len(child1.Children))
	assert.Equal(t, child2.ID, child1.Children[0].ID)
	assert.Equal(t, 1, len(doc.Body))
}

func TestIndent2(t *testing.T) {
	doc := NewDocument("test document", ToDateTime(time.Now().UTC()))
	para1 := doc.InsertParagraph(0, "First paragraph")
	child1 := NewParagraph("Child paragraph 1")
	para1.AddChild(child1)
	child2 := NewParagraph("Child paragraph 2")
	para1.AddChild(child2)

	_, _, err := doc.Indent(child1.ID)
	assert.Error(t, err)
	if err == nil {
		t.Fatal("Expected error when indenting first child, got nil")
	}
}

func TestUnindentRoot(t *testing.T) {
	doc := NewDocument("test document", ToDateTime(time.Now().UTC()))
	_ = doc.InsertParagraph(0, "First paragraph")
	para2 := doc.InsertParagraph(1, "Second paragraph")

	_, _, err := doc.UnIndent(para2.ID)
	assert.Error(t, err)
	if err == nil {
		t.Fatal("Expected error when unindenting top-level paragraph, got nil")
	}
}

func TestUnindent(t *testing.T) {
	doc := NewDocument("test document", ToDateTime(time.Now().UTC()))
	para1 := doc.InsertParagraph(0, "First paragraph")
	child1 := NewParagraph("Child paragraph 1")
	para1.AddChild(child1)

	paragraph, parent, err := doc.UnIndent(child1.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	assert.Equal(t, 0, len(para1.Children))
	assert.Equal(t, 2, len(doc.Body))
	assert.Equal(t, child1.ID, paragraph.ID)
	assert.Nil(t, parent)
}
