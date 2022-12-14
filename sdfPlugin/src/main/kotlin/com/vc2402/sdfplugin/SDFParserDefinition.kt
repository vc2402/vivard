package com.vc2402.sdfplugin

// Copyright 2000-2022 JetBrains s.r.o. and other contributors. Use of this source code is governed by the Apache 2.0 license that can be found in the LICENSE file.

import com.intellij.lang.ASTNode
import com.intellij.lang.ParserDefinition
import com.intellij.lang.PsiParser
import com.intellij.lexer.Lexer
import com.intellij.openapi.project.Project
import com.intellij.psi.FileViewProvider
import com.intellij.psi.PsiElement
import com.intellij.psi.PsiFile
import com.intellij.psi.tree.IFileElementType
import com.intellij.psi.tree.TokenSet
import com.vc2402.sdfplugin.parser.Parser
import com.vc2402.sdfplugin.psi.TokenSets
import com.vc2402.sdfplugin.psi.Types
import com.vc2402.sdfplugin.psi.SDFPsiFile


class SDFParserDefinition : ParserDefinition {
    override fun createLexer(project: Project): Lexer {
        return LexerAdapter()
    }

    override fun getCommentTokens(): TokenSet {
        return TokenSets.COMMENTS
    }

    override fun getStringLiteralElements(): TokenSet {
        return TokenSet.EMPTY
    }

    override fun createParser(project: Project): PsiParser {
        return Parser()
    }

    override fun getFileNodeType(): IFileElementType {
        return FILE
    }

    override fun createFile(viewProvider: FileViewProvider): PsiFile {
        return SDFPsiFile(viewProvider)
    }

    override fun createElement(node: ASTNode): PsiElement {
        return Types.Factory.createElement(node)
    }

    companion object {
        val FILE = IFileElementType(SDFLanguage.INSTANCE)
    }
}