package main

import (
	"github.com/expr-lang/expr/ast"
)

type FindExternalRefsVisitor struct {
	externalRefs []string
}

func (v *FindExternalRefsVisitor) Visit(node *ast.Node) {
	var ok bool
	var callNode *ast.CallNode
	var identifierNode *ast.IdentifierNode
	var stringNode *ast.StringNode

	if callNode, ok = (*node).(*ast.CallNode); ok && len(callNode.Arguments) > 0 && callNode.Callee != nil {
		if identifierNode, ok = callNode.Callee.(*ast.IdentifierNode); ok && identifierNode.Value == "external_ref" {
			if stringNode, ok = callNode.Arguments[0].(*ast.StringNode); ok {
				v.externalRefs = append(v.externalRefs, stringNode.Value)
			}
		}
	}
}
