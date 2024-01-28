// Copyright 2024 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bleveutils

import (
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
)

// DocType returns the name of a variable type.
func DocType(doc any) string {
	return reflect.Indirect(reflect.ValueOf(doc)).Type().Name()
}

// GetIndexMapping is a callback for index mapping allocation.
type GetIndexMapping func(im *mapping.IndexMappingImpl) *mapping.IndexMappingImpl

// GetDocumentMapping is a callback which can return an alternative
// DocumentMapping for typ. To use the default mapping return dm.
// To disable mapping return nil.
type GetDocumentMapping func(typ reflect.Type, dm *mapping.DocumentMapping) *mapping.DocumentMapping

// GetFieldMapping is a callback which can return an alternative FieldMapping
// for typ. To use the default mapping return fm. To disable mapping return nil.
type GetFieldMapping func(typ reflect.Type, fm *mapping.FieldMapping) *mapping.FieldMapping

// Build builds a new search index at indexPath from documents and calls dmcb
// for each document mapping to be added and fmcb for each field mapping to be
// added. It returns an open bleve.Index or an error.
//
// For more details see Builder.
func Build(indexPath string, imcb GetIndexMapping, dmcb GetDocumentMapping, fmcb GetFieldMapping, documents ...any) (idx bleve.Index, err error) {
	var (
		builder = Builder{imcb, dmcb, fmcb}
		mapping mapping.IndexMapping
	)
	if mapping, err = builder.BuildIndexMapping(documents...); err != nil {
		return nil, err
	}
	return bleve.New(indexPath, mapping)
}

// Builder builds bleve index mappings from documents.
//
// See typeToMapping for rules on how field mapping types are determined from
// Go types.
type Builder struct {
	imcb GetIndexMapping
	dmcb GetDocumentMapping
	fmcb GetFieldMapping
}

// NewBuilder returns a new Builder with a callback for document and field
// mapping allocation. Both are optional and can be nil.
func NewBuilder(imcb GetIndexMapping, dmcb GetDocumentMapping, fmcb GetFieldMapping) *Builder {
	return &Builder{imcb, dmcb, fmcb}
}

// BuildIndexMapping builds a bleve index mapping from documents.
func (self *Builder) BuildIndexMapping(documents ...any) (m mapping.IndexMapping, err error) {

	// Check for duplicate type names.
	var names = make(map[string]struct{})
	for _, doc := range documents {
		var doctype = DocType(doc)
		if _, exists := names[doctype]; exists {
			return nil, errors.New("duplicate document type: " + doctype)
		}
		names[doctype] = struct{}{}
	}

	// Build index mapping.
	var indexMapping = bleve.NewIndexMapping()
	if self.imcb != nil {
		indexMapping = self.imcb(indexMapping)
	}
	for _, doc := range documents {
		var docMapping *mapping.DocumentMapping
		if docMapping, err = self.buildDocumentMapping(doc); err != nil {
			return nil, err
		}
		indexMapping.AddDocumentMapping(DocType(doc), docMapping)
	}

	return indexMapping, nil
}

// buildDocumentMapping builds a bleve document mapping from doc.
func (self *Builder) buildDocumentMapping(doc any) (docMapping *mapping.DocumentMapping, err error) {

	var v = reflect.Indirect(reflect.ValueOf(doc))
	if v.Kind() != reflect.Struct {
		return nil, errors.New("document must be a struct")
	}

	docMapping = bleve.NewDocumentStaticMapping()
	if self.dmcb != nil {
		docMapping = self.dmcb(v.Type(), docMapping)
	}
	self.buildFieldMappings("", doc, v, docMapping)

	return docMapping, nil
}

// buildFieldMappings processes doc struct fields and adds field mappings to m under
// optionaly prefix prefixed name, dot separated, and the field name which is
// parsed from json tag first, field name second. Unexported fields are skipped.
func (self *Builder) buildFieldMappings(prefix string, doc any, v reflect.Value, docMapping *mapping.DocumentMapping) {
	for i := 0; i < v.NumField(); i++ {
		// Get name from field name.
		var name = v.Type().Field(i).Name
		// Uppercase/exported only.
		if name == "_" || (name[0] >= 97 && name[0] <= 122) {
			continue
		}
		// Get name from json tag.
		if jtag, exists := v.Type().Field(i).Tag.Lookup("json"); exists {
			if left, _, _ := strings.Cut(jtag, ","); left != "" && left != "-" {
				name = left
			}
		}
		var (
			typ          = v.Type().Field(i).Type
			fieldMapping = self.typeToMapping(typ)
		)
		if typ.Kind() == reflect.Struct && fieldMapping == nil {
			var dm = mapping.NewDocumentStaticMapping()
			if self.dmcb != nil {
				dm = self.dmcb(typ, dm)
			}
			self.buildFieldMappings(name, doc, v.Field(i), dm)
			docMapping.AddSubDocumentMapping(name, dm)
		}
		if fieldMapping != nil {
			// fm.Store = false
			if prefix != "" {
				name = prefix + "." + name
			}
			docMapping.AddFieldMappingsAt(name, fieldMapping)
		}
	}
}

// typeToMapping returns a bleve field mapping based on field typ.
func (self *Builder) typeToMapping(typ reflect.Type) (fieldMapping *mapping.FieldMapping) {
	switch typ.Kind() {
	case reflect.Bool:
		fieldMapping = bleve.NewBooleanFieldMapping()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr, reflect.Float32, reflect.Float64:
		fieldMapping = bleve.NewNumericFieldMapping()
	case reflect.String:
		fieldMapping = bleve.NewTextFieldMapping()
	case reflect.Struct:
		if typ.AssignableTo(timeType) {
			fieldMapping = bleve.NewDateTimeFieldMapping()
		}
	case reflect.Array, reflect.Slice:
		return self.typeToMapping(typ.Elem())
	case reflect.Map:
		return self.typeToMapping(typ.Elem())
	}
	if fieldMapping != nil && self.fmcb != nil {
		fieldMapping = self.fmcb(typ, fieldMapping)
	}
	return
}

var (
	timeType = reflect.TypeOf(time.Now())
)
