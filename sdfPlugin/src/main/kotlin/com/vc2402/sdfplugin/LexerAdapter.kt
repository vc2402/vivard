package com.vc2402.sdfplugin

import com.intellij.lexer.FlexAdapter


class LexerAdapter : FlexAdapter(SDFLexer(null))