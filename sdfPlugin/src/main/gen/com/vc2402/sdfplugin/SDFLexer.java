/* The following code was generated by JFlex 1.7.0 tweaked for IntelliJ platform */

package com.vc2402.sdfplugin;

import com.intellij.lexer.FlexLexer;
import com.intellij.psi.tree.IElementType;
import com.vc2402.sdfplugin.psi.Types;
import com.intellij.psi.TokenType;


/**
 * This class is a scanner generated by 
 * <a href="http://www.jflex.de/">JFlex</a> 1.7.0
 * from the specification file <tt>sdf.flex</tt>
 */
class SDFLexer implements FlexLexer {

  /** This character denotes the end of file */
  public static final int YYEOF = -1;

  /** initial size of the lookahead buffer */
  private static final int ZZ_BUFFERSIZE = 16384;

  /** lexical states */
  public static final int YYINITIAL = 0;
  public static final int WAITING_TYPE = 2;
  public static final int WAITING_ATTR_MODIFIER = 4;

  /**
   * ZZ_LEXSTATE[l] is the state in the DFA for the lexical state l
   * ZZ_LEXSTATE[l+1] is the state in the DFA for the lexical state l
   *                  at the beginning of a line
   * l is of the form l = 2*k, k a non negative integer
   */
  private static final int ZZ_LEXSTATE[] = { 
     0,  0,  1,  1,  2, 2
  };

  /** 
   * Translates characters to character classes
   * Chosen bits are [9, 6, 6]
   * Total runtime size is 1824 bytes
   */
  public static int ZZ_CMAP(int ch) {
    return ZZ_CMAP_A[(ZZ_CMAP_Y[ZZ_CMAP_Z[ch>>12]|((ch>>6)&0x3f)]<<6)|(ch&0x3f)];
  }

  /* The ZZ_CMAP_Z table has 272 entries */
  static final char ZZ_CMAP_Z[] = zzUnpackCMap(
    "\1\0\1\100\1\200\14\100\1\300\u0100\100");

  /* The ZZ_CMAP_Y table has 256 entries */
  static final char ZZ_CMAP_Y[] = zzUnpackCMap(
    "\1\0\1\1\1\2\175\3\1\4\176\3\1\5");

  /* The ZZ_CMAP_A table has 384 entries */
  static final char ZZ_CMAP_A[] = zzUnpackCMap(
    "\1\67\10\0\1\66\1\64\1\63\1\65\1\2\22\0\1\65\1\62\1\46\1\11\1\41\3\0\1\50"+
    "\1\51\2\0\1\61\1\34\1\35\1\1\12\40\1\42\1\36\1\56\1\60\1\57\1\0\1\43\3\37"+
    "\1\3\4\37\1\7\3\37\1\5\15\37\1\52\1\47\1\53\1\0\1\6\1\0\1\13\1\26\1\14\1\10"+
    "\1\17\1\31\1\16\1\44\1\32\1\37\1\15\1\33\1\23\1\22\1\30\1\12\1\37\1\27\1\25"+
    "\1\20\1\4\1\45\1\37\1\24\1\21\1\37\1\54\1\0\1\55\7\0\1\63\240\0\1\70\1\0\2"+
    "\63\125\0\1\71");

  /** 
   * Translates DFA states to action switch labels.
   */
  private static final int [] ZZ_ACTION = zzUnpackAction();

  private static final String ZZ_ACTION_PACKED_0 =
    "\3\0\2\1\1\2\2\3\1\4\10\3\1\5\1\6"+
    "\1\5\1\1\1\7\2\1\1\10\1\11\1\12\1\13"+
    "\1\14\1\15\1\16\1\17\1\20\1\21\1\22\1\4"+
    "\10\23\1\24\1\25\13\23\1\26\3\1\1\27\1\30"+
    "\1\31\1\32\1\0\14\3\1\5\1\33\5\0\1\34"+
    "\1\0\17\23\1\35\1\23\1\36\5\0\1\37\1\0"+
    "\1\40\15\3\1\41\10\0\2\23\1\42\3\23\1\43"+
    "\11\23\10\0\4\3\1\44\2\3\1\45\1\46\1\3"+
    "\1\47\1\3\1\33\3\0\1\50\3\0\1\51\1\52"+
    "\1\23\1\53\4\23\1\54\2\0\1\23\1\36\3\0"+
    "\1\55\3\0\11\3\4\0\1\23\1\56\3\23\2\0"+
    "\1\23\4\0\3\3\1\57\4\3\1\0\1\50\1\0"+
    "\1\60\1\61\2\23\3\0\1\55\1\0\1\3\1\62"+
    "\2\3\1\63\2\3\1\50\1\23\2\0\1\55\2\3"+
    "\2\0\2\3\5\0\1\35";

  private static int [] zzUnpackAction() {
    int [] result = new int[268];
    int offset = 0;
    offset = zzUnpackAction(ZZ_ACTION_PACKED_0, offset, result);
    return result;
  }

  private static int zzUnpackAction(String packed, int offset, int [] result) {
    int i = 0;       /* index in packed string  */
    int j = offset;  /* index in unpacked array */
    int l = packed.length();
    while (i < l) {
      int count = packed.charAt(i++);
      int value = packed.charAt(i++);
      do result[j++] = value; while (--count > 0);
    }
    return j;
  }


  /** 
   * Translates a state to a row index in the transition table
   */
  private static final int [] ZZ_ROWMAP = zzUnpackRowMap();

  private static final String ZZ_ROWMAP_PACKED_0 =
    "\0\0\0\72\0\164\0\256\0\350\0\u0122\0\u015c\0\u0196"+
    "\0\u01d0\0\u020a\0\u0244\0\u027e\0\u02b8\0\u02f2\0\u032c\0\u0366"+
    "\0\u03a0\0\u03da\0\256\0\u0414\0\u044e\0\256\0\u0488\0\u04c2"+
    "\0\256\0\256\0\256\0\256\0\256\0\256\0\256\0\256"+
    "\0\256\0\256\0\u0122\0\u04fc\0\u0536\0\u0570\0\u05aa\0\u05e4"+
    "\0\u061e\0\u0658\0\u0692\0\u06cc\0\256\0\256\0\u0706\0\u0740"+
    "\0\u077a\0\u07b4\0\u07ee\0\u0828\0\u0862\0\u089c\0\u08d6\0\u0910"+
    "\0\u094a\0\u0984\0\u09be\0\u09f8\0\u0a32\0\256\0\256\0\256"+
    "\0\u0a6c\0\u0aa6\0\u0ae0\0\u0b1a\0\u0b54\0\u0b8e\0\u0bc8\0\u0c02"+
    "\0\u0c3c\0\u0c76\0\u0cb0\0\u0cea\0\u0d24\0\u0d5e\0\u0d98\0\u0dd2"+
    "\0\u0e0c\0\u0e46\0\u0e80\0\u0eba\0\u04c2\0\256\0\u0ef4\0\u0f2e"+
    "\0\u0f68\0\u0fa2\0\u0fdc\0\u1016\0\u1050\0\u108a\0\u10c4\0\u10fe"+
    "\0\u1138\0\u1172\0\u11ac\0\u11e6\0\u1220\0\u125a\0\u0740\0\u1294"+
    "\0\u12ce\0\u1308\0\u1342\0\u137c\0\u13b6\0\u0a32\0\256\0\u13f0"+
    "\0\u142a\0\u1464\0\u149e\0\u14d8\0\u1512\0\u154c\0\u1586\0\u15c0"+
    "\0\u15fa\0\u1634\0\u166e\0\u16a8\0\u16e2\0\u171c\0\u0414\0\u1756"+
    "\0\u1790\0\u17ca\0\u1804\0\u183e\0\u1878\0\u18b2\0\u18ec\0\u1926"+
    "\0\u1960\0\u0536\0\u199a\0\u19d4\0\u1a0e\0\u0536\0\u1a48\0\u1a82"+
    "\0\u1abc\0\u1af6\0\u1b30\0\u1b6a\0\u1ba4\0\u1bde\0\u1c18\0\u1c52"+
    "\0\u1c8c\0\u1cc6\0\u1d00\0\u1d3a\0\u1d74\0\u1dae\0\u1de8\0\u1e22"+
    "\0\u1e5c\0\u1e96\0\u1ed0\0\u015c\0\u1f0a\0\u1f44\0\u015c\0\u015c"+
    "\0\u1f7e\0\u015c\0\u1fb8\0\u1ff2\0\u202c\0\u2066\0\u20a0\0\u20da"+
    "\0\u2114\0\u214e\0\u2188\0\u0536\0\u0536\0\u21c2\0\u0536\0\u21fc"+
    "\0\u2236\0\u2270\0\u22aa\0\u0740\0\u22e4\0\u231e\0\u2358\0\u2392"+
    "\0\u23cc\0\u2406\0\u2440\0\u247a\0\u24b4\0\u24ee\0\u2528\0\u2562"+
    "\0\u259c\0\u25d6\0\u2610\0\u264a\0\u2684\0\u26be\0\u26f8\0\u2732"+
    "\0\u276c\0\u27a6\0\u27e0\0\u281a\0\u2854\0\u0536\0\u288e\0\u28c8"+
    "\0\u2902\0\u293c\0\u2976\0\u29b0\0\u29ea\0\u2a24\0\u2a5e\0\u2a98"+
    "\0\u2ad2\0\u2b0c\0\u2b46\0\u015c\0\u2b80\0\u2bba\0\u2bf4\0\u2c2e"+
    "\0\u2c68\0\u2ca2\0\u2cdc\0\u0536\0\u0740\0\u2d16\0\u2d50\0\u2d8a"+
    "\0\u2dc4\0\u2dfe\0\u2e38\0\u2e72\0\u2eac\0\u015c\0\u2ee6\0\u2f20"+
    "\0\u015c\0\u2f5a\0\u2f94\0\u2fce\0\u3008\0\u3042\0\u307c\0\u30b6"+
    "\0\u30f0\0\u312a\0\u3164\0\u319e\0\u31d8\0\u3212\0\u324c\0\u3286"+
    "\0\u32c0\0\u32fa\0\u3334\0\256";

  private static int [] zzUnpackRowMap() {
    int [] result = new int[268];
    int offset = 0;
    offset = zzUnpackRowMap(ZZ_ROWMAP_PACKED_0, offset, result);
    return result;
  }

  private static int zzUnpackRowMap(String packed, int offset, int [] result) {
    int i = 0;  /* index in packed string  */
    int j = offset;  /* index in unpacked array */
    int l = packed.length();
    while (i < l) {
      int high = packed.charAt(i++) << 16;
      result[j++] = high | packed.charAt(i++);
    }
    return j;
  }

  /** 
   * The transition table of the DFA
   */
  private static final int [] ZZ_TRANS = zzUnpackTrans();

  private static final String ZZ_TRANS_PACKED_0 =
    "\1\4\1\5\1\6\5\7\1\10\1\11\1\12\1\13"+
    "\1\14\2\7\1\15\1\16\2\7\1\17\1\7\1\20"+
    "\3\7\1\21\2\7\1\4\1\22\1\23\1\7\1\24"+
    "\1\25\1\26\1\27\2\7\1\30\1\4\1\31\1\32"+
    "\1\33\1\34\1\35\1\36\1\37\1\4\1\40\1\41"+
    "\1\42\1\6\2\43\1\44\5\4\1\6\5\45\1\46"+
    "\1\4\1\45\1\47\7\45\1\50\1\45\1\51\1\52"+
    "\2\45\1\53\1\54\1\45\2\4\1\23\1\45\4\4"+
    "\2\45\3\4\1\32\1\55\1\56\2\4\1\37\2\4"+
    "\1\41\1\42\4\6\5\4\1\6\1\57\5\60\1\4"+
    "\1\60\1\61\1\62\2\60\1\63\1\64\6\60\1\65"+
    "\1\66\1\67\1\70\1\71\1\4\1\72\1\4\1\60"+
    "\1\72\1\73\1\4\1\74\2\60\1\75\1\4\1\76"+
    "\1\77\5\4\1\100\1\40\1\41\1\4\4\6\3\4"+
    "\73\0\1\101\72\0\1\6\60\0\4\6\6\0\6\7"+
    "\1\0\22\7\1\0\1\102\1\0\2\7\3\0\2\7"+
    "\27\0\6\7\1\0\20\7\1\103\1\7\1\0\1\102"+
    "\1\0\2\7\3\0\2\7\24\0\2\11\1\0\61\11"+
    "\1\0\5\11\3\0\6\7\1\0\1\7\1\104\20\7"+
    "\1\0\1\102\1\0\2\7\3\0\2\7\27\0\6\7"+
    "\1\0\14\7\1\105\5\7\1\0\1\102\1\0\2\7"+
    "\3\0\2\7\27\0\6\7\1\0\16\7\1\106\3\7"+
    "\1\0\1\102\1\0\2\7\3\0\2\7\27\0\6\7"+
    "\1\0\10\7\1\107\1\110\1\111\7\7\1\0\1\102"+
    "\1\0\2\7\3\0\2\7\27\0\6\7\1\0\7\7"+
    "\1\112\5\7\1\113\4\7\1\0\1\102\1\0\2\7"+
    "\3\0\2\7\27\0\6\7\1\0\5\7\1\114\14\7"+
    "\1\0\1\102\1\0\2\7\3\0\2\7\27\0\6\7"+
    "\1\0\20\7\1\115\1\7\1\0\1\102\1\0\2\7"+
    "\3\0\2\7\27\0\6\7\1\0\1\7\1\116\20\7"+
    "\1\0\1\102\1\0\2\7\3\0\2\7\61\0\1\117"+
    "\2\0\1\24\66\0\1\24\2\0\1\24\34\0\6\120"+
    "\1\0\22\120\3\0\1\120\4\0\2\120\40\0\1\121"+
    "\3\0\1\122\4\0\1\123\1\0\1\124\42\0\46\125"+
    "\1\126\1\127\22\125\2\11\1\6\60\11\1\44\1\6"+
    "\2\44\3\11\3\0\6\45\1\0\22\45\1\0\1\102"+
    "\1\0\2\45\3\0\2\45\27\0\6\45\1\0\1\45"+
    "\1\130\20\45\1\0\1\102\1\0\2\45\3\0\2\45"+
    "\27\0\1\45\1\131\4\45\1\0\22\45\1\0\1\102"+
    "\1\0\2\45\3\0\2\45\27\0\6\45\1\0\1\45"+
    "\1\132\20\45\1\0\1\102\1\0\2\45\3\0\2\45"+
    "\27\0\6\45\1\0\6\45\1\133\13\45\1\0\1\102"+
    "\1\0\2\45\3\0\2\45\27\0\6\45\1\0\16\45"+
    "\1\134\3\45\1\0\1\102\1\0\2\45\3\0\2\45"+
    "\27\0\6\45\1\0\21\45\1\135\1\0\1\102\1\0"+
    "\2\45\3\0\2\45\27\0\6\45\1\0\10\45\1\136"+
    "\11\45\1\0\1\102\1\0\2\45\3\0\2\45\27\0"+
    "\1\60\1\137\4\60\1\0\22\60\3\0\2\60\3\0"+
    "\2\60\27\0\6\60\1\0\22\60\3\0\2\60\3\0"+
    "\2\60\27\0\1\60\1\140\4\60\1\0\22\60\3\0"+
    "\2\60\3\0\2\60\27\0\6\60\1\0\1\60\1\141"+
    "\20\60\3\0\2\60\3\0\2\60\27\0\6\60\1\0"+
    "\11\60\1\142\10\60\3\0\2\60\3\0\2\60\27\0"+
    "\6\60\1\0\15\60\1\143\4\60\3\0\2\60\3\0"+
    "\2\60\27\0\6\60\1\0\5\60\1\144\14\60\3\0"+
    "\2\60\3\0\2\60\27\0\6\60\1\0\10\60\1\145"+
    "\11\60\3\0\2\60\3\0\2\60\27\0\6\60\1\0"+
    "\1\60\1\146\20\60\3\0\2\60\3\0\2\60\27\0"+
    "\5\60\1\147\1\0\22\60\3\0\2\60\3\0\2\60"+
    "\27\0\6\60\1\0\16\60\1\150\3\60\3\0\2\60"+
    "\3\0\2\60\61\0\1\72\2\0\1\72\34\0\6\151"+
    "\1\0\22\151\3\0\1\151\4\0\2\151\40\0\1\152"+
    "\3\0\1\153\4\0\1\154\1\0\1\155\42\0\46\156"+
    "\1\157\1\160\22\156\2\101\1\0\61\101\1\0\5\101"+
    "\3\0\6\161\1\0\22\161\3\0\1\161\4\0\2\161"+
    "\27\0\6\7\1\0\2\7\1\162\17\7\1\0\1\102"+
    "\1\0\2\7\3\0\2\7\27\0\6\7\1\0\2\7"+
    "\1\163\17\7\1\0\1\102\1\0\2\7\3\0\2\7"+
    "\27\0\6\7\1\0\13\7\1\164\6\7\1\0\1\102"+
    "\1\0\2\7\3\0\2\7\27\0\6\7\1\0\10\7"+
    "\1\165\11\7\1\0\1\102\1\0\2\7\3\0\2\7"+
    "\27\0\1\7\1\166\4\7\1\0\22\7\1\0\1\102"+
    "\1\0\2\7\3\0\2\7\27\0\6\7\1\0\14\7"+
    "\1\167\5\7\1\0\1\102\1\0\2\7\3\0\2\7"+
    "\27\0\6\7\1\0\6\7\1\170\13\7\1\0\1\102"+
    "\1\0\2\7\3\0\2\7\27\0\6\7\1\0\1\171"+
    "\21\7\1\0\1\102\1\0\2\7\3\0\2\7\27\0"+
    "\1\7\1\172\4\7\1\0\1\7\1\173\20\7\1\0"+
    "\1\102\1\0\2\7\3\0\2\7\27\0\6\7\1\0"+
    "\6\7\1\174\13\7\1\0\1\102\1\0\2\7\3\0"+
    "\2\7\27\0\6\7\1\0\10\7\1\175\11\7\1\0"+
    "\1\102\1\0\2\7\3\0\2\7\27\0\6\7\1\0"+
    "\21\7\1\176\1\0\1\102\1\0\2\7\3\0\2\7"+
    "\61\0\1\177\2\0\1\24\34\0\6\120\1\0\23\120"+
    "\2\0\2\120\1\0\1\200\1\0\2\120\53\0\1\201"+
    "\14\0\1\202\57\0\1\203\56\0\1\204\1\205\70\0"+
    "\1\206\141\0\1\207\5\0\6\45\1\0\6\45\1\210"+
    "\13\45\1\0\1\102\1\0\2\45\3\0\2\45\27\0"+
    "\6\45\1\0\6\45\1\211\13\45\1\0\1\102\1\0"+
    "\2\45\3\0\2\45\27\0\6\45\1\0\1\212\21\45"+
    "\1\0\1\102\1\0\2\45\3\0\2\45\27\0\6\45"+
    "\1\0\15\45\1\213\4\45\1\0\1\102\1\0\2\45"+
    "\3\0\2\45\27\0\6\45\1\0\16\45\1\214\3\45"+
    "\1\0\1\102\1\0\2\45\3\0\2\45\27\0\6\45"+
    "\1\0\16\45\1\215\3\45\1\0\1\102\1\0\2\45"+
    "\3\0\2\45\27\0\6\45\1\0\6\45\1\216\13\45"+
    "\1\0\1\102\1\0\2\45\3\0\2\45\27\0\2\60"+
    "\1\217\3\60\1\0\22\60\3\0\2\60\3\0\2\60"+
    "\27\0\6\60\1\0\6\60\1\220\13\60\3\0\2\60"+
    "\3\0\2\60\27\0\6\60\1\0\21\60\1\221\3\0"+
    "\2\60\3\0\2\60\27\0\6\60\1\0\14\60\1\222"+
    "\5\60\3\0\2\60\3\0\2\60\27\0\1\60\1\223"+
    "\4\60\1\0\22\60\3\0\2\60\3\0\2\60\27\0"+
    "\6\60\1\0\17\60\1\224\2\60\3\0\2\60\3\0"+
    "\2\60\27\0\6\60\1\0\5\60\1\225\14\60\3\0"+
    "\2\60\3\0\2\60\27\0\6\60\1\0\21\60\1\226"+
    "\3\0\2\60\3\0\2\60\27\0\6\60\1\0\16\60"+
    "\1\227\3\60\3\0\2\60\3\0\2\60\27\0\6\151"+
    "\1\0\23\151\2\0\2\151\1\0\1\230\1\0\2\151"+
    "\53\0\1\231\14\0\1\232\57\0\1\233\56\0\1\234"+
    "\1\235\70\0\1\236\141\0\1\237\5\0\6\161\1\0"+
    "\22\161\3\0\2\161\3\0\2\161\27\0\6\7\1\0"+
    "\6\7\1\240\13\7\1\0\1\102\1\0\2\7\3\0"+
    "\2\7\27\0\6\7\1\0\3\7\1\241\16\7\1\0"+
    "\1\102\1\0\2\7\3\0\2\7\27\0\6\7\1\0"+
    "\6\7\1\242\13\7\1\0\1\102\1\0\2\7\3\0"+
    "\2\7\27\0\6\7\1\0\17\7\1\243\2\7\1\0"+
    "\1\102\1\0\2\7\3\0\2\7\27\0\6\7\1\0"+
    "\11\7\1\244\10\7\1\0\1\102\1\0\2\7\3\0"+
    "\2\7\27\0\6\7\1\0\5\7\1\245\14\7\1\0"+
    "\1\102\1\0\2\7\3\0\2\7\27\0\6\7\1\0"+
    "\5\7\1\246\14\7\1\0\1\102\1\0\2\7\3\0"+
    "\2\7\27\0\6\7\1\0\5\7\1\247\14\7\1\0"+
    "\1\102\1\0\2\7\3\0\2\7\27\0\6\7\1\0"+
    "\5\7\1\250\14\7\1\0\1\102\1\0\2\7\3\0"+
    "\2\7\27\0\6\7\1\0\10\7\1\251\11\7\1\0"+
    "\1\102\1\0\2\7\3\0\2\7\27\0\6\7\1\0"+
    "\1\7\1\252\20\7\1\0\1\102\1\0\2\7\3\0"+
    "\2\7\27\0\6\7\1\0\4\7\1\253\15\7\1\0"+
    "\1\102\1\0\2\7\3\0\2\7\27\0\6\7\1\0"+
    "\13\7\1\172\6\7\1\0\1\102\1\0\2\7\3\0"+
    "\2\7\27\0\6\254\1\0\22\254\3\0\1\254\4\0"+
    "\2\254\43\0\1\255\65\0\1\256\101\0\1\257\66\0"+
    "\1\260\64\0\1\261\103\0\1\262\134\0\1\263\4\0"+
    "\6\45\1\0\5\45\1\264\14\45\1\0\1\102\1\0"+
    "\2\45\3\0\2\45\27\0\6\45\1\0\16\45\1\265"+
    "\3\45\1\0\1\102\1\0\2\45\3\0\2\45\27\0"+
    "\6\45\1\0\20\45\1\266\1\45\1\0\1\102\1\0"+
    "\2\45\3\0\2\45\27\0\6\45\1\0\21\45\1\267"+
    "\1\0\1\102\1\0\2\45\3\0\2\45\27\0\6\45"+
    "\1\0\1\45\1\270\20\45\1\0\1\102\1\0\2\45"+
    "\3\0\2\45\27\0\3\60\1\271\2\60\1\0\22\60"+
    "\3\0\2\60\3\0\2\60\27\0\6\60\1\0\16\60"+
    "\1\147\3\60\3\0\2\60\3\0\2\60\27\0\6\60"+
    "\1\0\2\60\1\272\17\60\3\0\2\60\3\0\2\60"+
    "\27\0\6\60\1\0\5\60\1\273\14\60\3\0\2\60"+
    "\3\0\2\60\27\0\6\60\1\0\5\60\1\274\14\60"+
    "\3\0\2\60\3\0\2\60\27\0\6\60\1\0\22\60"+
    "\1\275\2\0\2\60\3\0\2\60\27\0\6\60\1\0"+
    "\22\60\1\276\2\0\2\60\3\0\2\60\27\0\6\60"+
    "\1\0\13\60\1\223\6\60\3\0\2\60\3\0\2\60"+
    "\27\0\6\60\1\0\3\60\1\277\16\60\3\0\2\60"+
    "\3\0\2\60\27\0\6\300\1\0\22\300\3\0\1\300"+
    "\4\0\2\300\43\0\1\301\65\0\1\302\101\0\1\303"+
    "\66\0\1\304\64\0\1\305\103\0\1\306\134\0\1\307"+
    "\4\0\6\7\1\0\20\7\1\310\1\7\1\0\1\102"+
    "\1\0\2\7\3\0\2\7\27\0\6\7\1\0\1\7"+
    "\1\311\20\7\1\0\1\102\1\0\2\7\3\0\2\7"+
    "\27\0\6\7\1\0\15\7\1\312\4\7\1\0\1\102"+
    "\1\0\2\7\3\0\2\7\27\0\6\7\1\0\20\7"+
    "\1\313\1\7\1\0\1\102\1\0\2\7\3\0\2\7"+
    "\27\0\5\7\1\314\1\0\22\7\1\0\1\102\1\0"+
    "\2\7\3\0\2\7\27\0\6\7\1\0\10\7\1\315"+
    "\4\7\1\316\4\7\1\0\1\102\1\0\2\7\3\0"+
    "\2\7\27\0\6\7\1\0\13\7\1\317\6\7\1\0"+
    "\1\102\1\0\2\7\3\0\2\7\27\0\6\7\1\0"+
    "\21\7\1\320\1\0\1\102\1\0\2\7\3\0\2\7"+
    "\27\0\6\254\1\0\23\254\2\0\2\254\3\0\2\254"+
    "\37\0\1\321\100\0\1\322\66\0\1\260\114\0\1\323"+
    "\56\0\1\204\72\0\1\324\132\0\1\125\3\0\6\45"+
    "\1\0\10\45\1\325\11\45\1\0\1\102\1\0\2\45"+
    "\3\0\2\45\27\0\6\45\1\0\6\45\1\326\13\45"+
    "\1\0\1\102\1\0\2\45\3\0\2\45\27\0\4\60"+
    "\1\327\1\60\1\0\22\60\3\0\2\60\3\0\2\60"+
    "\27\0\1\60\1\330\4\60\1\0\22\60\3\0\2\60"+
    "\3\0\2\60\27\0\5\60\1\331\1\0\22\60\3\0"+
    "\2\60\3\0\2\60\43\0\1\332\72\0\1\333\54\0"+
    "\1\60\1\334\4\60\1\0\22\60\3\0\2\60\3\0"+
    "\2\60\27\0\6\300\1\0\23\300\2\0\2\300\3\0"+
    "\2\300\37\0\1\335\100\0\1\336\66\0\1\304\114\0"+
    "\1\337\56\0\1\234\72\0\1\340\132\0\1\156\3\0"+
    "\6\7\1\0\16\7\1\341\3\7\1\0\1\102\1\0"+
    "\2\7\3\0\2\7\27\0\6\7\1\0\4\7\1\342"+
    "\15\7\1\0\1\102\1\0\2\7\3\0\2\7\27\0"+
    "\6\7\1\0\1\7\1\343\20\7\1\0\1\102\1\0"+
    "\2\7\3\0\2\7\27\0\6\7\1\0\4\7\1\344"+
    "\15\7\1\0\1\102\1\0\2\7\3\0\2\7\27\0"+
    "\5\7\1\345\1\0\22\7\1\0\1\102\1\0\2\7"+
    "\3\0\2\7\27\0\5\7\1\346\1\0\22\7\1\0"+
    "\1\102\1\0\2\7\3\0\2\7\27\0\6\7\1\0"+
    "\10\7\1\344\11\7\1\0\1\102\1\0\2\7\3\0"+
    "\2\7\27\0\6\7\1\0\20\7\1\347\1\7\1\0"+
    "\1\102\1\0\2\7\3\0\2\7\27\0\6\7\1\0"+
    "\5\7\1\350\14\7\1\0\1\102\1\0\2\7\3\0"+
    "\2\7\44\0\1\257\67\0\1\351\56\0\6\352\1\0"+
    "\22\352\3\0\1\352\4\0\2\352\57\0\1\353\41\0"+
    "\6\45\1\0\4\45\1\354\15\45\1\0\1\102\1\0"+
    "\2\45\3\0\2\45\27\0\5\60\1\355\1\0\22\60"+
    "\3\0\2\60\3\0\2\60\27\0\6\60\1\0\21\60"+
    "\1\356\3\0\2\60\3\0\2\60\27\0\5\60\1\357"+
    "\1\0\22\60\3\0\2\60\3\0\2\60\47\0\1\360"+
    "\76\0\1\361\44\0\6\60\1\0\1\147\21\60\3\0"+
    "\2\60\3\0\2\60\44\0\1\303\67\0\1\362\56\0"+
    "\6\363\1\0\22\363\3\0\1\363\4\0\2\363\57\0"+
    "\1\364\41\0\6\7\1\0\10\7\1\365\11\7\1\0"+
    "\1\102\1\0\2\7\3\0\2\7\27\0\6\7\1\0"+
    "\5\7\1\366\14\7\1\0\1\102\1\0\2\7\3\0"+
    "\2\7\27\0\6\7\1\0\2\7\1\367\17\7\1\0"+
    "\1\102\1\0\2\7\3\0\2\7\27\0\6\7\1\0"+
    "\1\7\1\370\20\7\1\0\1\102\1\0\2\7\3\0"+
    "\2\7\27\0\6\7\1\0\1\7\1\370\11\7\1\371"+
    "\6\7\1\0\1\102\1\0\2\7\3\0\2\7\27\0"+
    "\6\7\1\0\5\7\1\372\14\7\1\0\1\102\1\0"+
    "\2\7\3\0\2\7\27\0\6\7\1\0\6\7\1\373"+
    "\13\7\1\0\1\102\1\0\2\7\3\0\2\7\43\0"+
    "\1\374\55\0\6\352\1\0\22\352\3\0\2\352\3\0"+
    "\2\352\71\0\1\257\27\0\6\60\1\0\1\60\1\375"+
    "\20\60\3\0\2\60\3\0\2\60\27\0\6\60\1\0"+
    "\5\60\1\70\14\60\3\0\2\60\3\0\2\60\52\0"+
    "\1\376\77\0\1\377\54\0\1\u0100\55\0\6\363\1\0"+
    "\22\363\3\0\2\363\3\0\2\363\71\0\1\303\27\0"+
    "\6\7\1\0\1\7\1\u0101\20\7\1\0\1\102\1\0"+
    "\2\7\3\0\2\7\27\0\6\7\1\0\6\7\1\344"+
    "\13\7\1\0\1\102\1\0\2\7\3\0\2\7\27\0"+
    "\6\7\1\0\14\7\1\u0102\5\7\1\0\1\102\1\0"+
    "\2\7\3\0\2\7\27\0\6\7\1\0\10\7\1\367"+
    "\11\7\1\0\1\102\1\0\2\7\3\0\2\7\27\0"+
    "\6\7\1\0\16\7\1\316\3\7\1\0\1\102\1\0"+
    "\2\7\3\0\2\7\34\0\1\260\31\0\1\323\32\0"+
    "\6\60\1\0\6\60\1\357\13\60\3\0\2\60\3\0"+
    "\2\60\43\0\1\u0103\75\0\1\u0104\56\0\1\304\31\0"+
    "\1\337\32\0\6\7\1\0\15\7\1\u0105\4\7\1\0"+
    "\1\102\1\0\2\7\3\0\2\7\27\0\6\7\1\0"+
    "\21\7\1\u0106\1\0\1\102\1\0\2\7\3\0\2\7"+
    "\34\0\1\u0107\74\0\1\u0108\61\0\6\7\1\0\7\7"+
    "\1\344\12\7\1\0\1\102\1\0\2\7\3\0\2\7"+
    "\27\0\6\7\1\0\5\7\1\344\14\7\1\0\1\102"+
    "\1\0\2\7\3\0\2\7\34\0\1\u0109\103\0\1\u010a"+
    "\66\0\1\u010b\73\0\1\u010c\60\0\1\u010c\61\0";

  private static int [] zzUnpackTrans() {
    int [] result = new int[13166];
    int offset = 0;
    offset = zzUnpackTrans(ZZ_TRANS_PACKED_0, offset, result);
    return result;
  }

  private static int zzUnpackTrans(String packed, int offset, int [] result) {
    int i = 0;       /* index in packed string  */
    int j = offset;  /* index in unpacked array */
    int l = packed.length();
    while (i < l) {
      int count = packed.charAt(i++);
      int value = packed.charAt(i++);
      value--;
      do result[j++] = value; while (--count > 0);
    }
    return j;
  }


  /* error codes */
  private static final int ZZ_UNKNOWN_ERROR = 0;
  private static final int ZZ_NO_MATCH = 1;
  private static final int ZZ_PUSHBACK_2BIG = 2;

  /* error messages for the codes above */
  private static final String[] ZZ_ERROR_MSG = {
    "Unknown internal scanner error",
    "Error: could not match input",
    "Error: pushback value was too large"
  };

  /**
   * ZZ_ATTRIBUTE[aState] contains the attributes of state <code>aState</code>
   */
  private static final int [] ZZ_ATTRIBUTE = zzUnpackAttribute();

  private static final String ZZ_ATTRIBUTE_PACKED_0 =
    "\3\0\1\11\16\1\1\11\2\1\1\11\2\1\12\11"+
    "\12\1\2\11\17\1\3\11\1\1\1\0\16\1\5\0"+
    "\1\11\1\0\22\1\5\0\1\11\1\0\17\1\10\0"+
    "\20\1\10\0\15\1\3\0\1\1\3\0\11\1\2\0"+
    "\2\1\3\0\1\1\3\0\11\1\4\0\5\1\2\0"+
    "\1\1\4\0\10\1\1\0\1\1\1\0\4\1\3\0"+
    "\1\1\1\0\11\1\2\0\3\1\2\0\2\1\5\0"+
    "\1\11";

  private static int [] zzUnpackAttribute() {
    int [] result = new int[268];
    int offset = 0;
    offset = zzUnpackAttribute(ZZ_ATTRIBUTE_PACKED_0, offset, result);
    return result;
  }

  private static int zzUnpackAttribute(String packed, int offset, int [] result) {
    int i = 0;       /* index in packed string  */
    int j = offset;  /* index in unpacked array */
    int l = packed.length();
    while (i < l) {
      int count = packed.charAt(i++);
      int value = packed.charAt(i++);
      do result[j++] = value; while (--count > 0);
    }
    return j;
  }

  /** the input device */
  private java.io.Reader zzReader;

  /** the current state of the DFA */
  private int zzState;

  /** the current lexical state */
  private int zzLexicalState = YYINITIAL;

  /** this buffer contains the current text to be matched and is
      the source of the yytext() string */
  private CharSequence zzBuffer = "";

  /** the textposition at the last accepting state */
  private int zzMarkedPos;

  /** the current text position in the buffer */
  private int zzCurrentPos;

  /** startRead marks the beginning of the yytext() string in the buffer */
  private int zzStartRead;

  /** endRead marks the last character in the buffer, that has been read
      from input */
  private int zzEndRead;

  /**
   * zzAtBOL == true <=> the scanner is currently at the beginning of a line
   */
  private boolean zzAtBOL = true;

  /** zzAtEOF == true <=> the scanner is at the EOF */
  private boolean zzAtEOF;

  /** denotes if the user-EOF-code has already been executed */
  private boolean zzEOFDone;


  /**
   * Creates a new scanner
   *
   * @param   in  the java.io.Reader to read input from.
   */
  SDFLexer(java.io.Reader in) {
    this.zzReader = in;
  }


  /** 
   * Unpacks the compressed character translation table.
   *
   * @param packed   the packed character translation table
   * @return         the unpacked character translation table
   */
  private static char [] zzUnpackCMap(String packed) {
    int size = 0;
    for (int i = 0, length = packed.length(); i < length; i += 2) {
      size += packed.charAt(i);
    }
    char[] map = new char[size];
    int i = 0;  /* index in packed string  */
    int j = 0;  /* index in unpacked array */
    while (i < packed.length()) {
      int  count = packed.charAt(i++);
      char value = packed.charAt(i++);
      do map[j++] = value; while (--count > 0);
    }
    return map;
  }

  public final int getTokenStart() {
    return zzStartRead;
  }

  public final int getTokenEnd() {
    return getTokenStart() + yylength();
  }

  public void reset(CharSequence buffer, int start, int end, int initialState) {
    zzBuffer = buffer;
    zzCurrentPos = zzMarkedPos = zzStartRead = start;
    zzAtEOF  = false;
    zzAtBOL = true;
    zzEndRead = end;
    yybegin(initialState);
  }

  /**
   * Refills the input buffer.
   *
   * @return      {@code false}, iff there was new input.
   *
   * @exception   java.io.IOException  if any I/O-Error occurs
   */
  private boolean zzRefill() throws java.io.IOException {
    return true;
  }


  /**
   * Returns the current lexical state.
   */
  public final int yystate() {
    return zzLexicalState;
  }


  /**
   * Enters a new lexical state
   *
   * @param newState the new lexical state
   */
  public final void yybegin(int newState) {
    zzLexicalState = newState;
  }


  /**
   * Returns the text matched by the current regular expression.
   */
  public final CharSequence yytext() {
    return zzBuffer.subSequence(zzStartRead, zzMarkedPos);
  }


  /**
   * Returns the character at position {@code pos} from the
   * matched text.
   *
   * It is equivalent to yytext().charAt(pos), but faster
   *
   * @param pos the position of the character to fetch.
   *            A value from 0 to yylength()-1.
   *
   * @return the character at position pos
   */
  public final char yycharat(int pos) {
    return zzBuffer.charAt(zzStartRead+pos);
  }


  /**
   * Returns the length of the matched text region.
   */
  public final int yylength() {
    return zzMarkedPos-zzStartRead;
  }


  /**
   * Reports an error that occurred while scanning.
   *
   * In a wellformed scanner (no or only correct usage of
   * yypushback(int) and a match-all fallback rule) this method
   * will only be called with things that "Can't Possibly Happen".
   * If this method is called, something is seriously wrong
   * (e.g. a JFlex bug producing a faulty scanner etc.).
   *
   * Usual syntax/scanner level error handling should be done
   * in error fallback rules.
   *
   * @param   errorCode  the code of the errormessage to display
   */
  private void zzScanError(int errorCode) {
    String message;
    try {
      message = ZZ_ERROR_MSG[errorCode];
    }
    catch (ArrayIndexOutOfBoundsException e) {
      message = ZZ_ERROR_MSG[ZZ_UNKNOWN_ERROR];
    }

    throw new Error(message);
  }


  /**
   * Pushes the specified amount of characters back into the input stream.
   *
   * They will be read again by then next call of the scanning method
   *
   * @param number  the number of characters to be read again.
   *                This number must not be greater than yylength()!
   */
  public void yypushback(int number)  {
    if ( number > yylength() )
      zzScanError(ZZ_PUSHBACK_2BIG);

    zzMarkedPos -= number;
  }


  /**
   * Contains user EOF-code, which will be executed exactly once,
   * when the end of file is reached
   */
  private void zzDoEOF() {
    if (!zzEOFDone) {
      zzEOFDone = true;
    
    }
  }


  /**
   * Resumes scanning until the next regular expression is matched,
   * the end of input is encountered or an I/O-Error occurs.
   *
   * @return      the next token
   * @exception   java.io.IOException  if any I/O-Error occurs
   */
  public IElementType advance() throws java.io.IOException {
    int zzInput;
    int zzAction;

    // cached fields:
    int zzCurrentPosL;
    int zzMarkedPosL;
    int zzEndReadL = zzEndRead;
    CharSequence zzBufferL = zzBuffer;

    int [] zzTransL = ZZ_TRANS;
    int [] zzRowMapL = ZZ_ROWMAP;
    int [] zzAttrL = ZZ_ATTRIBUTE;

    while (true) {
      zzMarkedPosL = zzMarkedPos;

      zzAction = -1;

      zzCurrentPosL = zzCurrentPos = zzStartRead = zzMarkedPosL;

      zzState = ZZ_LEXSTATE[zzLexicalState];

      // set up zzAction for empty match case:
      int zzAttributes = zzAttrL[zzState];
      if ( (zzAttributes & 1) == 1 ) {
        zzAction = zzState;
      }


      zzForAction: {
        while (true) {

          if (zzCurrentPosL < zzEndReadL) {
            zzInput = Character.codePointAt(zzBufferL, zzCurrentPosL/*, zzEndReadL*/);
            zzCurrentPosL += Character.charCount(zzInput);
          }
          else if (zzAtEOF) {
            zzInput = YYEOF;
            break zzForAction;
          }
          else {
            // store back cached positions
            zzCurrentPos  = zzCurrentPosL;
            zzMarkedPos   = zzMarkedPosL;
            boolean eof = zzRefill();
            // get translated positions and possibly new buffer
            zzCurrentPosL  = zzCurrentPos;
            zzMarkedPosL   = zzMarkedPos;
            zzBufferL      = zzBuffer;
            zzEndReadL     = zzEndRead;
            if (eof) {
              zzInput = YYEOF;
              break zzForAction;
            }
            else {
              zzInput = Character.codePointAt(zzBufferL, zzCurrentPosL/*, zzEndReadL*/);
              zzCurrentPosL += Character.charCount(zzInput);
            }
          }
          int zzNext = zzTransL[ zzRowMapL[zzState] + ZZ_CMAP(zzInput) ];
          if (zzNext == -1) break zzForAction;
          zzState = zzNext;

          zzAttributes = zzAttrL[zzState];
          if ( (zzAttributes & 1) == 1 ) {
            zzAction = zzState;
            zzMarkedPosL = zzCurrentPosL;
            if ( (zzAttributes & 8) == 8 ) break zzForAction;
          }

        }
      }

      // store back cached position
      zzMarkedPos = zzMarkedPosL;

      if (zzInput == YYEOF && zzStartRead == zzCurrentPos) {
        zzAtEOF = true;
        zzDoEOF();
        return null;
      }
      else {
        switch (zzAction < 0 ? zzAction : ZZ_ACTION[zzAction]) {
          case 1: 
            { return TokenType.BAD_CHARACTER;
            } 
            // fall through
          case 52: break;
          case 2: 
            { return TokenType.WHITE_SPACE;
            } 
            // fall through
          case 53: break;
          case 3: 
            { yybegin(YYINITIAL); return Types.IDENTIFIER;
            } 
            // fall through
          case 54: break;
          case 4: 
            { yybegin(YYINITIAL); return Types.META_LINE;
            } 
            // fall through
          case 55: break;
          case 5: 
            { yybegin(YYINITIAL); return Types.NUMBER_VALUE;
            } 
            // fall through
          case 56: break;
          case 6: 
            { yybegin(YYINITIAL); return Types.STATEMENT_END;
            } 
            // fall through
          case 57: break;
          case 7: 
            { yybegin(WAITING_TYPE); return Types.SEMI;
            } 
            // fall through
          case 58: break;
          case 8: 
            { yybegin(YYINITIAL); return Types.BR_OPEN;
            } 
            // fall through
          case 59: break;
          case 9: 
            { yybegin(YYINITIAL); return Types.BR_CLOSE;
            } 
            // fall through
          case 60: break;
          case 10: 
            { yybegin(YYINITIAL); return Types.BRACKETOPEN;
            } 
            // fall through
          case 61: break;
          case 11: 
            { yybegin(YYINITIAL); return Types.BRACKETCLOSE;
            } 
            // fall through
          case 62: break;
          case 12: 
            { yybegin(YYINITIAL); return Types.BRACESOPEN;
            } 
            // fall through
          case 63: break;
          case 13: 
            { yybegin(YYINITIAL); return Types.BRACESCLOSE;
            } 
            // fall through
          case 64: break;
          case 14: 
            { yybegin(WAITING_ATTR_MODIFIER); return Types.MODIFIEROPEN;
            } 
            // fall through
          case 65: break;
          case 15: 
            { return Types.EQUAL;
            } 
            // fall through
          case 66: break;
          case 16: 
            { yybegin(YYINITIAL); return Types.COMMA;
            } 
            // fall through
          case 67: break;
          case 17: 
            { return Types.NOTNULL;
            } 
            // fall through
          case 68: break;
          case 18: 
            { yybegin(YYINITIAL); return TokenType.WHITE_SPACE;
            } 
            // fall through
          case 69: break;
          case 19: 
            { return Types.IDENTIFIER;
            } 
            // fall through
          case 70: break;
          case 20: 
            { return Types.BRACKETOPEN;
            } 
            // fall through
          case 71: break;
          case 21: 
            { return Types.BRACKETCLOSE;
            } 
            // fall through
          case 72: break;
          case 22: 
            { return Types.NUMBER_VALUE;
            } 
            // fall through
          case 73: break;
          case 23: 
            { return Types.BR_OPEN;
            } 
            // fall through
          case 74: break;
          case 24: 
            { return Types.BR_CLOSE;
            } 
            // fall through
          case 75: break;
          case 25: 
            { yybegin(YYINITIAL); return Types.MODIFIERCLOSE;
            } 
            // fall through
          case 76: break;
          case 26: 
            { yybegin(YYINITIAL); return Types.COMMENT_LINE;
            } 
            // fall through
          case 77: break;
          case 27: 
            { yybegin(YYINITIAL); return Types.ANNOTATIONTAG;
            } 
            // fall through
          case 78: break;
          case 28: 
            { yybegin(YYINITIAL); return Types.STRING_VALUE;
            } 
            // fall through
          case 79: break;
          case 29: 
            { return Types.ATTRMODIFIER;
            } 
            // fall through
          case 80: break;
          case 30: 
            { return Types.ANNOTATIONTAG;
            } 
            // fall through
          case 81: break;
          case 31: 
            { return Types.STRING_VALUE;
            } 
            // fall through
          case 82: break;
          case 32: 
            { return Types.QUALIFIEDNAME;
            } 
            // fall through
          case 83: break;
          case 33: 
            { return Types.MORE;
            } 
            // fall through
          case 84: break;
          case 34: 
            { return Types.MAP;
            } 
            // fall through
          case 85: break;
          case 35: 
            { return Types.INT;
            } 
            // fall through
          case 86: break;
          case 36: 
            { yybegin(YYINITIAL); return Types.KW_ENUM;
            } 
            // fall through
          case 87: break;
          case 37: 
            { yybegin(YYINITIAL); return Types.KW_TYPE;
            } 
            // fall through
          case 88: break;
          case 38: 
            { yybegin(YYINITIAL); return Types.BOOL_VALUE;
            } 
            // fall through
          case 89: break;
          case 39: 
            { yybegin(YYINITIAL); return Types.META;
            } 
            // fall through
          case 90: break;
          case 40: 
            { yybegin(YYINITIAL); return Types.HOOKTAG;
            } 
            // fall through
          case 91: break;
          case 41: 
            { return Types.DATE;
            } 
            // fall through
          case 92: break;
          case 42: 
            { return Types.AUTO;
            } 
            // fall through
          case 93: break;
          case 43: 
            { return Types.BOOL;
            } 
            // fall through
          case 94: break;
          case 44: 
            { return Types.BOOL_VALUE;
            } 
            // fall through
          case 95: break;
          case 45: 
            { return Types.HOOKTAG;
            } 
            // fall through
          case 96: break;
          case 46: 
            { return Types.FLOAT;
            } 
            // fall through
          case 97: break;
          case 47: 
            { yybegin(YYINITIAL); return Types.TYPEMODIFIER;
            } 
            // fall through
          case 98: break;
          case 48: 
            { return Types.STRING;
            } 
            // fall through
          case 99: break;
          case 49: 
            { return Types.DUMMYIDENTIFIER;
            } 
            // fall through
          case 100: break;
          case 50: 
            { yybegin(YYINITIAL); return Types.PACKAGE;
            } 
            // fall through
          case 101: break;
          case 51: 
            { yybegin(YYINITIAL); return Types.EXTENDS;
            } 
            // fall through
          case 102: break;
          default:
            zzScanError(ZZ_NO_MATCH);
          }
      }
    }
  }


}
