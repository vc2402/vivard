// Copyright 2000-2022 JetBrains s.r.o. and other contributors. Use of this source code is governed by the Apache 2.0 license that can be found in the LICENSE file.

package com.vc2402.sdfplugin;

import com.intellij.lang.Language;

class SDFLanguage: Language("vvf") {

    companion object {
        val INSTANCE = SDFLanguage()
    }

}