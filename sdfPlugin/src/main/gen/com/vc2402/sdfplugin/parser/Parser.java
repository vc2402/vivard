// This is a generated file. Not intended for manual editing.
package com.vc2402.sdfplugin.parser;

import com.intellij.lang.PsiBuilder;
import com.intellij.lang.PsiBuilder.Marker;
import static com.vc2402.sdfplugin.psi.Types.*;
import static com.intellij.lang.parser.GeneratedParserUtilBase.*;
import com.intellij.psi.tree.IElementType;
import com.intellij.lang.ASTNode;
import com.intellij.psi.tree.TokenSet;
import com.intellij.lang.PsiParser;
import com.intellij.lang.LightPsiParser;

@SuppressWarnings({"SimplifiableIfStatement", "UnusedAssignment"})
public class Parser implements PsiParser, LightPsiParser {

  public ASTNode parse(IElementType t, PsiBuilder b) {
    parseLight(t, b);
    return b.getTreeBuilt();
  }

  public void parseLight(IElementType t, PsiBuilder b) {
    boolean r;
    b = adapt_builder_(t, b, this, null);
    Marker m = enter_section_(b, 0, _COLLAPSE_, null);
    r = parse_root_(t, b);
    exit_section_(b, 0, m, t, r, true, TRUE_CONDITION);
  }

  protected boolean parse_root_(IElementType t, PsiBuilder b) {
    return parse_root_(t, b, 0);
  }

  static boolean parse_root_(IElementType t, PsiBuilder b, int l) {
    return sdfFile(b, l + 1);
  }

  /* ********************************************************** */
  // ann_param_name EQUAL ann_param_value
  public static boolean ann_param(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "ann_param")) return false;
    boolean r;
    Marker m = enter_section_(b, l, _NONE_, ANN_PARAM, "<ann param>");
    r = ann_param_name(b, l + 1);
    r = r && consumeToken(b, EQUAL);
    r = r && ann_param_value(b, l + 1);
    exit_section_(b, l, m, r, false, null);
    return r;
  }

  /* ********************************************************** */
  // IDENTIFIER | KW_TYPE | ATTRMODIFIER | TYPEMODIFIER
  public static boolean ann_param_name(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "ann_param_name")) return false;
    boolean r;
    Marker m = enter_section_(b, l, _NONE_, ANN_PARAM_NAME, "<ann param name>");
    r = consumeToken(b, IDENTIFIER);
    if (!r) r = consumeToken(b, KW_TYPE);
    if (!r) r = consumeToken(b, ATTRMODIFIER);
    if (!r) r = consumeToken(b, TYPEMODIFIER);
    exit_section_(b, l, m, r, false, null);
    return r;
  }

  /* ********************************************************** */
  // DUMMYIDENTIFIER | NUMBER_VALUE | STRING_VALUE | BOOL_VALUE
  public static boolean ann_param_value(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "ann_param_value")) return false;
    boolean r;
    Marker m = enter_section_(b, l, _NONE_, ANN_PARAM_VALUE, "<ann param value>");
    r = consumeToken(b, DUMMYIDENTIFIER);
    if (!r) r = consumeToken(b, NUMBER_VALUE);
    if (!r) r = consumeToken(b, STRING_VALUE);
    if (!r) r = consumeToken(b, BOOL_VALUE);
    exit_section_(b, l, m, r, false, null);
    return r;
  }

  /* ********************************************************** */
  // DUMMYIDENTIFIER? ANNOTATIONTAG ( BR_OPEN annotation_values BR_CLOSE )?
  public static boolean annotation(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "annotation")) return false;
    if (!nextTokenIs(b, "<annotation>", ANNOTATIONTAG, DUMMYIDENTIFIER)) return false;
    boolean r;
    Marker m = enter_section_(b, l, _NONE_, ANNOTATION, "<annotation>");
    r = annotation_0(b, l + 1);
    r = r && consumeToken(b, ANNOTATIONTAG);
    r = r && annotation_2(b, l + 1);
    exit_section_(b, l, m, r, false, null);
    return r;
  }

  // DUMMYIDENTIFIER?
  private static boolean annotation_0(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "annotation_0")) return false;
    consumeToken(b, DUMMYIDENTIFIER);
    return true;
  }

  // ( BR_OPEN annotation_values BR_CLOSE )?
  private static boolean annotation_2(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "annotation_2")) return false;
    annotation_2_0(b, l + 1);
    return true;
  }

  // BR_OPEN annotation_values BR_CLOSE
  private static boolean annotation_2_0(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "annotation_2_0")) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeToken(b, BR_OPEN);
    r = r && annotation_values(b, l + 1);
    r = r && consumeToken(b, BR_CLOSE);
    exit_section_(b, m, null, r);
    return r;
  }

  /* ********************************************************** */
  // ann_param
  //     | IDENTIFIER
  //     | STRING_VALUE
  //     | DUMMYIDENTIFIER
  public static boolean annotation_value(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "annotation_value")) return false;
    boolean r;
    Marker m = enter_section_(b, l, _NONE_, ANNOTATION_VALUE, "<annotation value>");
    r = ann_param(b, l + 1);
    if (!r) r = consumeToken(b, IDENTIFIER);
    if (!r) r = consumeToken(b, STRING_VALUE);
    if (!r) r = consumeToken(b, DUMMYIDENTIFIER);
    exit_section_(b, l, m, r, false, null);
    return r;
  }

  /* ********************************************************** */
  // annotation_value*
  public static boolean annotation_values(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "annotation_values")) return false;
    Marker m = enter_section_(b, l, _NONE_, ANNOTATION_VALUES, "<annotation values>");
    while (true) {
      int c = current_position_(b);
      if (!annotation_value(b, l + 1)) break;
      if (!empty_element_parsed_guard_(b, "annotation_values", c)) break;
    }
    exit_section_(b, l, m, true, false, null);
    return true;
  }

  /* ********************************************************** */
  // BRACKETOPEN DUMMYIDENTIFIER? type BRACKETCLOSE
  public static boolean array_type(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "array_type")) return false;
    if (!nextTokenIs(b, BRACKETOPEN)) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeToken(b, BRACKETOPEN);
    r = r && array_type_1(b, l + 1);
    r = r && type(b, l + 1);
    r = r && consumeToken(b, BRACKETCLOSE);
    exit_section_(b, m, ARRAY_TYPE, r);
    return r;
  }

  // DUMMYIDENTIFIER?
  private static boolean array_type_1(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "array_type_1")) return false;
    consumeToken(b, DUMMYIDENTIFIER);
    return true;
  }

  /* ********************************************************** */
  // DUMMYIDENTIFIER | HOOKTAG | ATTRMODIFIER | annotation
  public static boolean attr_modifier(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "attr_modifier")) return false;
    boolean r;
    Marker m = enter_section_(b, l, _NONE_, ATTR_MODIFIER, "<attr modifier>");
    r = consumeToken(b, DUMMYIDENTIFIER);
    if (!r) r = consumeToken(b, HOOKTAG);
    if (!r) r = consumeToken(b, ATTRMODIFIER);
    if (!r) r = annotation(b, l + 1);
    exit_section_(b, l, m, r, false, null);
    return r;
  }

  /* ********************************************************** */
  // attr_modifier*
  public static boolean attr_modifiers(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "attr_modifiers")) return false;
    Marker m = enter_section_(b, l, _NONE_, ATTR_MODIFIERS, "<attr modifiers>");
    while (true) {
      int c = current_position_(b);
      if (!attr_modifier(b, l + 1)) break;
      if (!empty_element_parsed_guard_(b, "attr_modifiers", c)) break;
    }
    exit_section_(b, l, m, true, false, null);
    return true;
  }

  /* ********************************************************** */
  // COMMENT_LINE
  public static boolean comment(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "comment")) return false;
    if (!nextTokenIs(b, COMMENT_LINE)) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeToken(b, COMMENT_LINE);
    exit_section_(b, m, COMMENT, r);
    return r;
  }

  /* ********************************************************** */
  // comment*
  public static boolean comments(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "comments")) return false;
    Marker m = enter_section_(b, l, _NONE_, COMMENTS, "<comments>");
    while (true) {
      int c = current_position_(b);
      if (!comment(b, l + 1)) break;
      if (!empty_element_parsed_guard_(b, "comments", c)) break;
    }
    exit_section_(b, l, m, true, false, null);
    return true;
  }

  /* ********************************************************** */
  // DUMMYIDENTIFIER
  //     | type_declaration
  //     | meta_declaration
  //     | comment
  public static boolean declaration(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "declaration")) return false;
    boolean r;
    Marker m = enter_section_(b, l, _NONE_, DECLARATION, "<declaration>");
    r = consumeToken(b, DUMMYIDENTIFIER);
    if (!r) r = type_declaration(b, l + 1);
    if (!r) r = meta_declaration(b, l + 1);
    if (!r) r = comment(b, l + 1);
    exit_section_(b, l, m, r, false, null);
    return r;
  }

  /* ********************************************************** */
  // declaration*
  public static boolean declarations(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "declarations")) return false;
    Marker m = enter_section_(b, l, _NONE_, DECLARATIONS, "<declarations>");
    while (true) {
      int c = current_position_(b);
      if (!declaration(b, l + 1)) break;
      if (!empty_element_parsed_guard_(b, "declarations", c)) break;
    }
    exit_section_(b, l, m, true, false, null);
    return true;
  }

  /* ********************************************************** */
  // entry*
  public static boolean entries(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "entries")) return false;
    Marker m = enter_section_(b, l, _NONE_, ENTRIES, "<entries>");
    while (true) {
      int c = current_position_(b);
      if (!entry(b, l + 1)) break;
      if (!empty_element_parsed_guard_(b, "entries", c)) break;
    }
    exit_section_(b, l, m, true, false, null);
    return true;
  }

  /* ********************************************************** */
  // (field | method) (MODIFIEROPEN attr_modifiers MODIFIERCLOSE)? STATEMENT_END
  public static boolean entry(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "entry")) return false;
    if (!nextTokenIs(b, IDENTIFIER)) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = entry_0(b, l + 1);
    r = r && entry_1(b, l + 1);
    r = r && consumeToken(b, STATEMENT_END);
    exit_section_(b, m, ENTRY, r);
    return r;
  }

  // field | method
  private static boolean entry_0(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "entry_0")) return false;
    boolean r;
    r = field(b, l + 1);
    if (!r) r = method(b, l + 1);
    return r;
  }

  // (MODIFIEROPEN attr_modifiers MODIFIERCLOSE)?
  private static boolean entry_1(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "entry_1")) return false;
    entry_1_0(b, l + 1);
    return true;
  }

  // MODIFIEROPEN attr_modifiers MODIFIERCLOSE
  private static boolean entry_1_0(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "entry_1_0")) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeToken(b, MODIFIEROPEN);
    r = r && attr_modifiers(b, l + 1);
    r = r && consumeToken(b, MODIFIERCLOSE);
    exit_section_(b, m, null, r);
    return r;
  }

  /* ********************************************************** */
  // IDENTIFIER (SEMI | DUMMYIDENTIFIER) (type | DUMMYIDENTIFIER)
  public static boolean field(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "field")) return false;
    if (!nextTokenIs(b, IDENTIFIER)) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeToken(b, IDENTIFIER);
    r = r && field_1(b, l + 1);
    r = r && field_2(b, l + 1);
    exit_section_(b, m, FIELD, r);
    return r;
  }

  // SEMI | DUMMYIDENTIFIER
  private static boolean field_1(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "field_1")) return false;
    boolean r;
    r = consumeToken(b, SEMI);
    if (!r) r = consumeToken(b, DUMMYIDENTIFIER);
    return r;
  }

  // type | DUMMYIDENTIFIER
  private static boolean field_2(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "field_2")) return false;
    boolean r;
    r = type(b, l + 1);
    if (!r) r = consumeToken(b, DUMMYIDENTIFIER);
    return r;
  }

  /* ********************************************************** */
  // HOOKTAG "=" STRING_VALUE
  //   | HOOKTAG
  public static boolean hook_tag(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "hook_tag")) return false;
    if (!nextTokenIs(b, HOOKTAG)) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = hook_tag_0(b, l + 1);
    if (!r) r = consumeToken(b, HOOKTAG);
    exit_section_(b, m, HOOK_TAG, r);
    return r;
  }

  // HOOKTAG "=" STRING_VALUE
  private static boolean hook_tag_0(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "hook_tag_0")) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeToken(b, HOOKTAG);
    r = r && consumeToken(b, "=");
    r = r && consumeToken(b, STRING_VALUE);
    exit_section_(b, m, null, r);
    return r;
  }

  /* ********************************************************** */
  // st_int | st_string
  public static boolean map_index_type(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "map_index_type")) return false;
    if (!nextTokenIs(b, "<map index type>", INT, STRING)) return false;
    boolean r;
    Marker m = enter_section_(b, l, _NONE_, MAP_INDEX_TYPE, "<map index type>");
    r = st_int(b, l + 1);
    if (!r) r = st_string(b, l + 1);
    exit_section_(b, l, m, r, false, null);
    return r;
  }

  /* ********************************************************** */
  // MAP BRACKETOPEN DUMMYIDENTIFIER? map_index_type BRACKETCLOSE DUMMYIDENTIFIER? type
  public static boolean map_type(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "map_type")) return false;
    if (!nextTokenIs(b, MAP)) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeTokens(b, 0, MAP, BRACKETOPEN);
    r = r && map_type_2(b, l + 1);
    r = r && map_index_type(b, l + 1);
    r = r && consumeToken(b, BRACKETCLOSE);
    r = r && map_type_5(b, l + 1);
    r = r && type(b, l + 1);
    exit_section_(b, m, MAP_TYPE, r);
    return r;
  }

  // DUMMYIDENTIFIER?
  private static boolean map_type_2(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "map_type_2")) return false;
    consumeToken(b, DUMMYIDENTIFIER);
    return true;
  }

  // DUMMYIDENTIFIER?
  private static boolean map_type_5(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "map_type_5")) return false;
    consumeToken(b, DUMMYIDENTIFIER);
    return true;
  }

  /* ********************************************************** */
  // META BR_OPEN IDENTIFIER BR_CLOSE
  //     META_LINE*
  public static boolean meta_declaration(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "meta_declaration")) return false;
    if (!nextTokenIs(b, META)) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeTokens(b, 0, META, BR_OPEN, IDENTIFIER, BR_CLOSE);
    r = r && meta_declaration_4(b, l + 1);
    exit_section_(b, m, META_DECLARATION, r);
    return r;
  }

  // META_LINE*
  private static boolean meta_declaration_4(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "meta_declaration_4")) return false;
    while (true) {
      int c = current_position_(b);
      if (!consumeToken(b, META_LINE)) break;
      if (!empty_element_parsed_guard_(b, "meta_declaration_4", c)) break;
    }
    return true;
  }

  /* ********************************************************** */
  // IDENTIFIER BR_OPEN params? BR_CLOSE (SEMI DUMMYIDENTIFIER? type)?
  public static boolean method(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "method")) return false;
    if (!nextTokenIs(b, IDENTIFIER)) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeTokens(b, 0, IDENTIFIER, BR_OPEN);
    r = r && method_2(b, l + 1);
    r = r && consumeToken(b, BR_CLOSE);
    r = r && method_4(b, l + 1);
    exit_section_(b, m, METHOD, r);
    return r;
  }

  // params?
  private static boolean method_2(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "method_2")) return false;
    params(b, l + 1);
    return true;
  }

  // (SEMI DUMMYIDENTIFIER? type)?
  private static boolean method_4(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "method_4")) return false;
    method_4_0(b, l + 1);
    return true;
  }

  // SEMI DUMMYIDENTIFIER? type
  private static boolean method_4_0(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "method_4_0")) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeToken(b, SEMI);
    r = r && method_4_0_1(b, l + 1);
    r = r && type(b, l + 1);
    exit_section_(b, m, null, r);
    return r;
  }

  // DUMMYIDENTIFIER?
  private static boolean method_4_0_1(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "method_4_0_1")) return false;
    consumeToken(b, DUMMYIDENTIFIER);
    return true;
  }

  /* ********************************************************** */
  // PACKAGE IDENTIFIER STATEMENT_END
  public static boolean package_declaration(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "package_declaration")) return false;
    if (!nextTokenIs(b, PACKAGE)) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeTokens(b, 0, PACKAGE, IDENTIFIER, STATEMENT_END);
    exit_section_(b, m, PACKAGE_DECLARATION, r);
    return r;
  }

  /* ********************************************************** */
  // IDENTIFIER SEMI DUMMYIDENTIFIER? type
  public static boolean param(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "param")) return false;
    if (!nextTokenIs(b, IDENTIFIER)) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeTokens(b, 0, IDENTIFIER, SEMI);
    r = r && param_2(b, l + 1);
    r = r && type(b, l + 1);
    exit_section_(b, m, PARAM, r);
    return r;
  }

  // DUMMYIDENTIFIER?
  private static boolean param_2(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "param_2")) return false;
    consumeToken(b, DUMMYIDENTIFIER);
    return true;
  }

  /* ********************************************************** */
  // param ( COMMA param)*
  public static boolean params(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "params")) return false;
    if (!nextTokenIs(b, IDENTIFIER)) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = param(b, l + 1);
    r = r && params_1(b, l + 1);
    exit_section_(b, m, PARAMS, r);
    return r;
  }

  // ( COMMA param)*
  private static boolean params_1(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "params_1")) return false;
    while (true) {
      int c = current_position_(b);
      if (!params_1_0(b, l + 1)) break;
      if (!empty_element_parsed_guard_(b, "params_1", c)) break;
    }
    return true;
  }

  // COMMA param
  private static boolean params_1_0(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "params_1_0")) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeToken(b, COMMA);
    r = r && param(b, l + 1);
    exit_section_(b, m, null, r);
    return r;
  }

  /* ********************************************************** */
  // DUMMYIDENTIFIER? comments? DUMMYIDENTIFIER? <package declaration>? declarations?
  static boolean sdfFile(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "sdfFile")) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = sdfFile_0(b, l + 1);
    r = r && sdfFile_1(b, l + 1);
    r = r && sdfFile_2(b, l + 1);
    r = r && sdfFile_3(b, l + 1);
    r = r && sdfFile_4(b, l + 1);
    exit_section_(b, m, null, r);
    return r;
  }

  // DUMMYIDENTIFIER?
  private static boolean sdfFile_0(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "sdfFile_0")) return false;
    consumeToken(b, DUMMYIDENTIFIER);
    return true;
  }

  // comments?
  private static boolean sdfFile_1(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "sdfFile_1")) return false;
    comments(b, l + 1);
    return true;
  }

  // DUMMYIDENTIFIER?
  private static boolean sdfFile_2(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "sdfFile_2")) return false;
    consumeToken(b, DUMMYIDENTIFIER);
    return true;
  }

  // <package declaration>?
  private static boolean sdfFile_3(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "sdfFile_3")) return false;
    package_declaration(b, l + 1);
    return true;
  }

  // declarations?
  private static boolean sdfFile_4(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "sdfFile_4")) return false;
    declarations(b, l + 1);
    return true;
  }

  /* ********************************************************** */
  // st_int | st_float | st_string | st_bool | st_date
  public static boolean simple_type(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "simple_type")) return false;
    boolean r;
    Marker m = enter_section_(b, l, _NONE_, SIMPLE_TYPE, "<simple type>");
    r = st_int(b, l + 1);
    if (!r) r = st_float(b, l + 1);
    if (!r) r = st_string(b, l + 1);
    if (!r) r = st_bool(b, l + 1);
    if (!r) r = st_date(b, l + 1);
    exit_section_(b, l, m, r, false, null);
    return r;
  }

  /* ********************************************************** */
  // BOOL
  public static boolean st_bool(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "st_bool")) return false;
    if (!nextTokenIs(b, BOOL)) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeToken(b, BOOL);
    exit_section_(b, m, ST_BOOL, r);
    return r;
  }

  /* ********************************************************** */
  // DATE
  public static boolean st_date(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "st_date")) return false;
    if (!nextTokenIs(b, DATE)) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeToken(b, DATE);
    exit_section_(b, m, ST_DATE, r);
    return r;
  }

  /* ********************************************************** */
  // FLOAT
  public static boolean st_float(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "st_float")) return false;
    if (!nextTokenIs(b, FLOAT)) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeToken(b, FLOAT);
    exit_section_(b, m, ST_FLOAT, r);
    return r;
  }

  /* ********************************************************** */
  // INT
  public static boolean st_int(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "st_int")) return false;
    if (!nextTokenIs(b, INT)) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeToken(b, INT);
    exit_section_(b, m, ST_INT, r);
    return r;
  }

  /* ********************************************************** */
  // STRING
  public static boolean st_string(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "st_string")) return false;
    if (!nextTokenIs(b, STRING)) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = consumeToken(b, STRING);
    exit_section_(b, m, ST_STRING, r);
    return r;
  }

  /* ********************************************************** */
  // (simple_type | IDENTIFIER | QUALIFIEDNAME | array_type | map_type) NOTNULL?
  public static boolean type(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "type")) return false;
    boolean r;
    Marker m = enter_section_(b, l, _NONE_, TYPE, "<type>");
    r = type_0(b, l + 1);
    r = r && type_1(b, l + 1);
    exit_section_(b, l, m, r, false, null);
    return r;
  }

  // simple_type | IDENTIFIER | QUALIFIEDNAME | array_type | map_type
  private static boolean type_0(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "type_0")) return false;
    boolean r;
    r = simple_type(b, l + 1);
    if (!r) r = consumeToken(b, IDENTIFIER);
    if (!r) r = consumeToken(b, QUALIFIEDNAME);
    if (!r) r = array_type(b, l + 1);
    if (!r) r = map_type(b, l + 1);
    return r;
  }

  // NOTNULL?
  private static boolean type_1(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "type_1")) return false;
    consumeToken(b, NOTNULL);
    return true;
  }

  /* ********************************************************** */
  // type_modifiers? (KW_TYPE IDENTIFIER | QUALIFIEDNAME) (DUMMYIDENTIFIER? EXTENDS DUMMYIDENTIFIER? IDENTIFIER | QUALIFIEDNAME)? DUMMYIDENTIFIER? BRACESOPEN DUMMYIDENTIFIER? entries MORE? BRACESCLOSE
  public static boolean type_declaration(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "type_declaration")) return false;
    boolean r;
    Marker m = enter_section_(b, l, _NONE_, TYPE_DECLARATION, "<type declaration>");
    r = type_declaration_0(b, l + 1);
    r = r && type_declaration_1(b, l + 1);
    r = r && type_declaration_2(b, l + 1);
    r = r && type_declaration_3(b, l + 1);
    r = r && consumeToken(b, BRACESOPEN);
    r = r && type_declaration_5(b, l + 1);
    r = r && entries(b, l + 1);
    r = r && type_declaration_7(b, l + 1);
    r = r && consumeToken(b, BRACESCLOSE);
    exit_section_(b, l, m, r, false, null);
    return r;
  }

  // type_modifiers?
  private static boolean type_declaration_0(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "type_declaration_0")) return false;
    type_modifiers(b, l + 1);
    return true;
  }

  // KW_TYPE IDENTIFIER | QUALIFIEDNAME
  private static boolean type_declaration_1(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "type_declaration_1")) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = parseTokens(b, 0, KW_TYPE, IDENTIFIER);
    if (!r) r = consumeToken(b, QUALIFIEDNAME);
    exit_section_(b, m, null, r);
    return r;
  }

  // (DUMMYIDENTIFIER? EXTENDS DUMMYIDENTIFIER? IDENTIFIER | QUALIFIEDNAME)?
  private static boolean type_declaration_2(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "type_declaration_2")) return false;
    type_declaration_2_0(b, l + 1);
    return true;
  }

  // DUMMYIDENTIFIER? EXTENDS DUMMYIDENTIFIER? IDENTIFIER | QUALIFIEDNAME
  private static boolean type_declaration_2_0(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "type_declaration_2_0")) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = type_declaration_2_0_0(b, l + 1);
    if (!r) r = consumeToken(b, QUALIFIEDNAME);
    exit_section_(b, m, null, r);
    return r;
  }

  // DUMMYIDENTIFIER? EXTENDS DUMMYIDENTIFIER? IDENTIFIER
  private static boolean type_declaration_2_0_0(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "type_declaration_2_0_0")) return false;
    boolean r;
    Marker m = enter_section_(b);
    r = type_declaration_2_0_0_0(b, l + 1);
    r = r && consumeToken(b, EXTENDS);
    r = r && type_declaration_2_0_0_2(b, l + 1);
    r = r && consumeToken(b, IDENTIFIER);
    exit_section_(b, m, null, r);
    return r;
  }

  // DUMMYIDENTIFIER?
  private static boolean type_declaration_2_0_0_0(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "type_declaration_2_0_0_0")) return false;
    consumeToken(b, DUMMYIDENTIFIER);
    return true;
  }

  // DUMMYIDENTIFIER?
  private static boolean type_declaration_2_0_0_2(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "type_declaration_2_0_0_2")) return false;
    consumeToken(b, DUMMYIDENTIFIER);
    return true;
  }

  // DUMMYIDENTIFIER?
  private static boolean type_declaration_3(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "type_declaration_3")) return false;
    consumeToken(b, DUMMYIDENTIFIER);
    return true;
  }

  // DUMMYIDENTIFIER?
  private static boolean type_declaration_5(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "type_declaration_5")) return false;
    consumeToken(b, DUMMYIDENTIFIER);
    return true;
  }

  // MORE?
  private static boolean type_declaration_7(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "type_declaration_7")) return false;
    consumeToken(b, MORE);
    return true;
  }

  /* ********************************************************** */
  // IDENTIFIER
  //     | TYPEMODIFIER
  //     | annotation
  //     | hook_tag
  public static boolean type_modifier(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "type_modifier")) return false;
    boolean r;
    Marker m = enter_section_(b, l, _NONE_, TYPE_MODIFIER, "<type modifier>");
    r = consumeToken(b, IDENTIFIER);
    if (!r) r = consumeToken(b, TYPEMODIFIER);
    if (!r) r = annotation(b, l + 1);
    if (!r) r = hook_tag(b, l + 1);
    exit_section_(b, l, m, r, false, null);
    return r;
  }

  /* ********************************************************** */
  // type_modifier*
  public static boolean type_modifiers(PsiBuilder b, int l) {
    if (!recursion_guard_(b, l, "type_modifiers")) return false;
    Marker m = enter_section_(b, l, _NONE_, TYPE_MODIFIERS, "<type modifiers>");
    while (true) {
      int c = current_position_(b);
      if (!type_modifier(b, l + 1)) break;
      if (!empty_element_parsed_guard_(b, "type_modifiers", c)) break;
    }
    exit_section_(b, l, m, true, false, null);
    return true;
  }

}
