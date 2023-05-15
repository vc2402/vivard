// Copyright 2000-2022 JetBrains s.r.o. and other contributors. Use of this source code is governed by the Apache 2.0 license that can be found in the LICENSE file.

package com.vc2402.sdfplugin;

import com.intellij.openapi.fileTypes.LanguageFileType;
import com.vc2402.sdfplugin.Icons

import javax.swing.*;

class SDFFileType: LanguageFileType(SDFLanguage.INSTANCE) {

    companion object {
        val INSTANCE = SDFFileType()
    }

    override fun getName() = "VVF"

    override fun getDescription() = "VVF language file"

    override fun getDefaultExtension() = "vvf"

    override fun getIcon() = Icons.FILE

}