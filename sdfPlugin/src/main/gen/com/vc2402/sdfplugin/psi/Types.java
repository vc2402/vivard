// This is a generated file. Not intended for manual editing.
package com.vc2402.sdfplugin.psi;

import com.intellij.psi.tree.IElementType;
import com.intellij.psi.PsiElement;
import com.intellij.lang.ASTNode;
import com.vc2402.sdfplugin.psi.impl.*;

public interface Types {

  IElementType ANNOTATION = new ElementType("ANNOTATION");
  IElementType ANNOTATION_VALUE = new ElementType("ANNOTATION_VALUE");
  IElementType ANNOTATION_VALUES = new ElementType("ANNOTATION_VALUES");
  IElementType ANN_PARAM = new ElementType("ANN_PARAM");
  IElementType ANN_PARAM_NAME = new ElementType("ANN_PARAM_NAME");
  IElementType ANN_PARAM_VALUE = new ElementType("ANN_PARAM_VALUE");
  IElementType ARRAY_TYPE = new ElementType("ARRAY_TYPE");
  IElementType ATTR_MODIFIER = new ElementType("ATTR_MODIFIER");
  IElementType ATTR_MODIFIERS = new ElementType("ATTR_MODIFIERS");
  IElementType COMMENT = new ElementType("COMMENT");
  IElementType COMMENTS = new ElementType("COMMENTS");
  IElementType DECLARATION = new ElementType("DECLARATION");
  IElementType DECLARATIONS = new ElementType("DECLARATIONS");
  IElementType ENTRIES = new ElementType("ENTRIES");
  IElementType ENTRY = new ElementType("ENTRY");
  IElementType FIELD = new ElementType("FIELD");
  IElementType HOOK_TAG = new ElementType("HOOK_TAG");
  IElementType MAP_INDEX_TYPE = new ElementType("MAP_INDEX_TYPE");
  IElementType MAP_TYPE = new ElementType("MAP_TYPE");
  IElementType META_DECLARATION = new ElementType("META_DECLARATION");
  IElementType METHOD = new ElementType("METHOD");
  IElementType PACKAGE_DECLARATION = new ElementType("PACKAGE_DECLARATION");
  IElementType PARAM = new ElementType("PARAM");
  IElementType PARAMS = new ElementType("PARAMS");
  IElementType SIMPLE_TYPE = new ElementType("SIMPLE_TYPE");
  IElementType ST_BOOL = new ElementType("ST_BOOL");
  IElementType ST_DATE = new ElementType("ST_DATE");
  IElementType ST_FLOAT = new ElementType("ST_FLOAT");
  IElementType ST_INT = new ElementType("ST_INT");
  IElementType ST_STRING = new ElementType("ST_STRING");
  IElementType TYPE = new ElementType("TYPE");
  IElementType TYPE_DECLARATION = new ElementType("TYPE_DECLARATION");
  IElementType TYPE_MODIFIER = new ElementType("TYPE_MODIFIER");
  IElementType TYPE_MODIFIERS = new ElementType("TYPE_MODIFIERS");

  IElementType ANNOTATIONTAG = new TokenType("ANNOTATIONTAG");
  IElementType ATTRMODIFIER = new TokenType("ATTRMODIFIER");
  IElementType BOOL = new TokenType("BOOL");
  IElementType BOOL_VALUE = new TokenType("BOOL_VALUE");
  IElementType BRACESCLOSE = new TokenType("BRACESCLOSE");
  IElementType BRACESOPEN = new TokenType("BRACESOPEN");
  IElementType BRACKETCLOSE = new TokenType("BRACKETCLOSE");
  IElementType BRACKETOPEN = new TokenType("BRACKETOPEN");
  IElementType BR_CLOSE = new TokenType("BR_CLOSE");
  IElementType BR_OPEN = new TokenType("BR_OPEN");
  IElementType COMMA = new TokenType("COMMA");
  IElementType COMMENT_LINE = new TokenType("COMMENT_LINE");
  IElementType DATE = new TokenType("DATE");
  IElementType DUMMYIDENTIFIER = new TokenType("DUMMYIDENTIFIER");
  IElementType EQUAL = new TokenType("EQUAL");
  IElementType EXTENDS = new TokenType("EXTENDS");
  IElementType FLOAT = new TokenType("FLOAT");
  IElementType HOOKTAG = new TokenType("HOOKTAG");
  IElementType IDENTIFIER = new TokenType("IDENTIFIER");
  IElementType INT = new TokenType("INT");
  IElementType KW_TYPE = new TokenType("KW_TYPE");
  IElementType MAP = new TokenType("MAP");
  IElementType META = new TokenType("META");
  IElementType META_LINE = new TokenType("META_LINE");
  IElementType MODIFIERCLOSE = new TokenType("MODIFIERCLOSE");
  IElementType MODIFIEROPEN = new TokenType("MODIFIEROPEN");
  IElementType MORE = new TokenType("MORE");
  IElementType NOTNULL = new TokenType("NOTNULL");
  IElementType NUMBER_VALUE = new TokenType("NUMBER_VALUE");
  IElementType PACKAGE = new TokenType("PACKAGE");
  IElementType QUALIFIEDNAME = new TokenType("QUALIFIEDNAME");
  IElementType SEMI = new TokenType("SEMI");
  IElementType STATEMENT_END = new TokenType("STATEMENT_END");
  IElementType STRING = new TokenType("STRING");
  IElementType STRING_VALUE = new TokenType("STRING_VALUE");
  IElementType TYPEMODIFIER = new TokenType("TYPEMODIFIER");

  class Factory {
    public static PsiElement createElement(ASTNode node) {
      IElementType type = node.getElementType();
      if (type == ANNOTATION) {
        return new SDFAnnotationImpl(node);
      }
      else if (type == ANNOTATION_VALUE) {
        return new SDFAnnotationValueImpl(node);
      }
      else if (type == ANNOTATION_VALUES) {
        return new SDFAnnotationValuesImpl(node);
      }
      else if (type == ANN_PARAM) {
        return new SDFAnnParamImpl(node);
      }
      else if (type == ANN_PARAM_NAME) {
        return new SDFAnnParamNameImpl(node);
      }
      else if (type == ANN_PARAM_VALUE) {
        return new SDFAnnParamValueImpl(node);
      }
      else if (type == ARRAY_TYPE) {
        return new SDFArrayTypeImpl(node);
      }
      else if (type == ATTR_MODIFIER) {
        return new SDFAttrModifierImpl(node);
      }
      else if (type == ATTR_MODIFIERS) {
        return new SDFAttrModifiersImpl(node);
      }
      else if (type == COMMENT) {
        return new SDFCommentImpl(node);
      }
      else if (type == COMMENTS) {
        return new SDFCommentsImpl(node);
      }
      else if (type == DECLARATION) {
        return new SDFDeclarationImpl(node);
      }
      else if (type == DECLARATIONS) {
        return new SDFDeclarationsImpl(node);
      }
      else if (type == ENTRIES) {
        return new SDFEntriesImpl(node);
      }
      else if (type == ENTRY) {
        return new SDFEntryImpl(node);
      }
      else if (type == FIELD) {
        return new SDFFieldImpl(node);
      }
      else if (type == HOOK_TAG) {
        return new SDFHookTagImpl(node);
      }
      else if (type == MAP_INDEX_TYPE) {
        return new SDFMapIndexTypeImpl(node);
      }
      else if (type == MAP_TYPE) {
        return new SDFMapTypeImpl(node);
      }
      else if (type == META_DECLARATION) {
        return new SDFMetaDeclarationImpl(node);
      }
      else if (type == METHOD) {
        return new SDFMethodImpl(node);
      }
      else if (type == PACKAGE_DECLARATION) {
        return new SDFPackageDeclarationImpl(node);
      }
      else if (type == PARAM) {
        return new SDFParamImpl(node);
      }
      else if (type == PARAMS) {
        return new SDFParamsImpl(node);
      }
      else if (type == SIMPLE_TYPE) {
        return new SDFSimpleTypeImpl(node);
      }
      else if (type == ST_BOOL) {
        return new SDFStBoolImpl(node);
      }
      else if (type == ST_DATE) {
        return new SDFStDateImpl(node);
      }
      else if (type == ST_FLOAT) {
        return new SDFStFloatImpl(node);
      }
      else if (type == ST_INT) {
        return new SDFStIntImpl(node);
      }
      else if (type == ST_STRING) {
        return new SDFStStringImpl(node);
      }
      else if (type == TYPE) {
        return new SDFTypeImpl(node);
      }
      else if (type == TYPE_DECLARATION) {
        return new SDFTypeDeclarationImpl(node);
      }
      else if (type == TYPE_MODIFIER) {
        return new SDFTypeModifierImpl(node);
      }
      else if (type == TYPE_MODIFIERS) {
        return new SDFTypeModifiersImpl(node);
      }
      throw new AssertionError("Unknown element type: " + type);
    }
  }
}
