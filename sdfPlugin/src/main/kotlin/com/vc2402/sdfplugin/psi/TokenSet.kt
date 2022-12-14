package com.vc2402.sdfplugin.psi

import com.intellij.psi.tree.TokenSet


interface TokenSets {
    companion object {
        val IDENTIFIERS = TokenSet.create(Types.TYPE)
        val COMMENTS = TokenSet.create(Types.COMMENT_LINE)
    }
}