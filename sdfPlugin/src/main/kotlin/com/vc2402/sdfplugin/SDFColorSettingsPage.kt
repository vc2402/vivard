package com.vc2402.sdfplugin

import com.intellij.openapi.editor.colors.TextAttributesKey
import com.intellij.openapi.fileTypes.SyntaxHighlighter
import com.intellij.openapi.options.colors.AttributesDescriptor
import com.intellij.openapi.options.colors.ColorDescriptor
import com.intellij.openapi.options.colors.ColorSettingsPage
import javax.swing.Icon

class SDFColorSettingsPage : ColorSettingsPage {
    override fun getIcon(): Icon? {
        return Icons.FILE
    }

    override fun getHighlighter(): SyntaxHighlighter {
        return SDFSyntaxHighlighter()
    }

    override fun getDemoText(): String {
        return """// You are reading the vivard description file.
package sample;

dictionary
type SomeDict {
  Id: int <auto>;
  Name: string <${"$"}js(title)>;
}

extendable
${"$"}vue-tabs(first second)
type Base {
  Id: string<id>;
  Name: string<${"$"}vue-tab(first name="Name" order=1 ignore=false)>;
}

@create
singleton
type Singleton {
  getSomeData(a:date):Data;
}
"""
    }

    override fun getAdditionalHighlightingTagToDescriptorMap(): Map<String, TextAttributesKey>? {
        return null
    }

    override fun getAttributeDescriptors(): Array<AttributesDescriptor> {
        return DESCRIPTORS
    }

    override fun getColorDescriptors(): Array<ColorDescriptor> {
        return ColorDescriptor.EMPTY_ARRAY
    }

    override fun getDisplayName(): String {
        return "sdf"
    }

    companion object {
        private val DESCRIPTORS = arrayOf(
            AttributesDescriptor("Keyword", SDFSyntaxHighlighter.KEY),
            AttributesDescriptor("Identifier", SDFSyntaxHighlighter.IDENT),
            AttributesDescriptor("Simple type", SDFSyntaxHighlighter.SIMPLE_TYPE),
            AttributesDescriptor("Modifier", SDFSyntaxHighlighter.MODIFIER),
            AttributesDescriptor("Annotation", SDFSyntaxHighlighter.ANNOTATION),
            AttributesDescriptor("Hook", SDFSyntaxHighlighter.HOOK),
            AttributesDescriptor("String value", SDFSyntaxHighlighter.STR_VAL),
            AttributesDescriptor("Number value", SDFSyntaxHighlighter.NUMB_VAL),
            AttributesDescriptor("Bool value", SDFSyntaxHighlighter.BOOL_VAL),
            AttributesDescriptor("Comment", SDFSyntaxHighlighter.COMMENT),
            AttributesDescriptor("Bad value", SDFSyntaxHighlighter.BAD_CHARACTER)
        )
    }
}