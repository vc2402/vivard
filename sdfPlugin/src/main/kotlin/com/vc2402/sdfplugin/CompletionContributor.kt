package com.vc2402.sdfplugin

import com.intellij.codeInsight.completion.*
import com.intellij.codeInsight.lookup.LookupElementBuilder
import com.intellij.openapi.project.DumbAware
import com.intellij.patterns.PatternCondition
import com.intellij.patterns.PlatformPatterns
import com.intellij.psi.PsiElement
import com.intellij.psi.PsiFile
import com.intellij.psi.TokenType
import com.intellij.psi.util.elementType
import com.intellij.util.ProcessingContext
import com.vc2402.sdfplugin.psi.SDFPsiFile
import com.vc2402.sdfplugin.psi.Types
import org.jetbrains.annotations.NotNull

class SDFCompletionContributor : CompletionContributor(), DumbAware {
    companion object {
        val MODIFIER_PATTERN = PlatformPatterns.psiElement().afterLeaf("@", "$")
        val DECLARATION_PATTERN = PlatformPatterns.or(
            PlatformPatterns.psiElement().withParent(PlatformPatterns.psiElement(Types.DECLARATION)),
            PlatformPatterns.psiElement().inside(PlatformPatterns.psiElement(Types.TYPE_MODIFIERS)),
            PlatformPatterns.psiElement().afterSiblingSkipping(PlatformPatterns.psiElement(TokenType.WHITE_SPACE), PlatformPatterns.psiElement(Types.TYPE_MODIFIER))
        )
        val ATTR_MODIFIER = PlatformPatterns.psiElement().inside(PlatformPatterns.psiElement(Types.ATTR_MODIFIERS))
        val TYPE_ANNOTATION_VALUES = PlatformPatterns.and(
            PlatformPatterns.psiElement().inside(PlatformPatterns.psiElement(Types.ANNOTATION_VALUE)),
            PlatformPatterns.psiElement().inside(PlatformPatterns.psiElement(Types.TYPE_MODIFIER))
        )
//            PlatformPatterns.psiElement()
//            .withParent(PlatformPatterns.psiElement(Types.ANNOTATION_VALUE))
//            .withParent(PlatformPatterns.psiElement(Types.ANNOTATION_VALUES))
//            .withParent(PlatformPatterns.psiElement(Types.ANNOTATION))
//            .withParent(PlatformPatterns.psiElement(Types.TYPE_MODIFIER))
        val TYPE_PATTERN = PlatformPatterns.psiElement()
            .afterLeafSkipping(PlatformPatterns.psiElement(Types.IDENTIFIER), PlatformPatterns.or(
                PlatformPatterns.psiElement(Types.SEMI), PlatformPatterns.psiElement(Types.BRACKETOPEN)))
    }
    init {
        extend(
            CompletionType.BASIC,
            DECLARATION_PATTERN,
            object : CompletionProvider<CompletionParameters>() {
                override fun addCompletions(
                    @NotNull parameters: CompletionParameters,
                    @NotNull context: ProcessingContext,
                    @NotNull resultSet: CompletionResultSet
                ) {
                    addTypeDeclarationCompletions(parameters, context, resultSet)
                }
            }
        )
        extend(
            CompletionType.BASIC,
            MODIFIER_PATTERN,
            object : CompletionProvider<CompletionParameters>() {
                override fun addCompletions(
                    @NotNull parameters: CompletionParameters,
                    @NotNull context: ProcessingContext,
                    @NotNull resultSet: CompletionResultSet
                ) {
                    var isType = true
                    var ps = parameters.position.prevSibling;
                    while(ps != null) {
                        if(ps.elementType == Types.MODIFIEROPEN) {
                            isType = false
                            break
                        }
                        if(ps.elementType == Types.DECLARATIONS || ps.elementType == Types.DECLARATION) {
                            break
                        }
                        ps = ps.prevSibling
                    }
                    var isAnnotation = parameters.offset > 0 && parameters.editor.document.charsSequence[parameters.offset - 1] == '$'
                    when {
                        isType && isAnnotation -> fillTypeAnnotations(resultSet, "")
                        isType && !isAnnotation -> fillTypeHooks(resultSet, "")
                        !isType && isAnnotation -> fillAttrAnnotations(resultSet, "")
                        else -> fillAttrHooks(resultSet, "")
                    }
                }
            }
        )

        //attr and method types completion
        extend(
            CompletionType.BASIC,
//            PlatformPatterns.psiElement()
//                .afterLeaf(":", "["),
            TYPE_PATTERN,
            object : CompletionProvider<CompletionParameters>() {
                override fun addCompletions(
                    @NotNull parameters: CompletionParameters,
                    @NotNull context: ProcessingContext,
                    @NotNull resultSet: CompletionResultSet
                ) {
                    addTypesComplition(parameters, context, resultSet)
                }
            }
        )
        extend(
            CompletionType.BASIC,
            ATTR_MODIFIER,
            object : CompletionProvider<CompletionParameters>() {
                override fun addCompletions(
                    @NotNull parameters: CompletionParameters,
                    @NotNull context: ProcessingContext,
                    @NotNull resultSet: CompletionResultSet
                ) {
                    var isAnnValues = false
                    var annotationTag = ""
                    var annParam: String? = null
                    var isMethod = false
                    var ps = parameters.position
                    while(ps != null) {
                        if(ps.elementType == Types.ANN_PARAM) {
                            if(ps.firstChild != null && ps.firstChild.elementType == Types.IDENTIFIER)
                                annParam = ps.firstChild.text
                        }
                        if(ps.elementType == Types.ANNOTATION_VALUES) {
                            isAnnValues = true
                            ps = ps.parent
                            if(ps != null && ps.elementType == Types.ANNOTATION) {
                                ps = ps.firstChild
                                if(ps != null && ps.elementType == Types.ANNOTATIONTAG)
                                    annotationTag = ps.text
                            }
                            break
                        }
                        if(ps.elementType == Types.ENTRY) {
                            if (ps.firstChild != null && ps.firstChild.elementType == Types.METHOD) {
                                isMethod = true
                            }
                            break
                        }
                        ps = ps.parent
                    }
                    if(isAnnValues) {
                        if(annParam != null)
                            fillValuesForTagParam(resultSet, annotationTag, annParam!!)
                        else
                            fillParamsForTag(resultSet, annotationTag)

                    } else {
                        resultSet.addElement(LookupElementBuilder.create("id"))
                        resultSet.addElement(LookupElementBuilder.create("auto"))
                        resultSet.addElement(LookupElementBuilder.create("lookup"))
                        resultSet.addElement(LookupElementBuilder.create("embedded"))
                        resultSet.addElement(LookupElementBuilder.create("calculated"))
                        resultSet.addElement(LookupElementBuilder.create("ref-embedded"))
                        resultSet.addElement(LookupElementBuilder.create("one-to-many"))
                        fillAttrHooks(resultSet, "@")
                        fillAttrAnnotations(resultSet, "$")
                    }
                }
            }
        )
        extend(
            CompletionType.BASIC,
            PlatformPatterns.psiElement().with(
                object: PatternCondition<PsiElement>("topLevel") {
                    override fun accepts(t: PsiElement, context: ProcessingContext?): Boolean {
//                        var p = t.parent
//                        if(p.elementType == Types.ANNOTATION_VALUE || p.elementType == Types.ANNOTATION_VALUES) {
//                            while(p != null) {
//                                if(p.elementType == Types.TYPE_MODIFIER)
//                                    return true
//                                p = p.parent
//                            }
//                        }
                        return true
                    }
                }
            ),
            object : CompletionProvider<CompletionParameters>() {
                override fun addCompletions(
                    @NotNull parameters: CompletionParameters,
                    @NotNull context: ProcessingContext,
                    @NotNull resultSet: CompletionResultSet
                ) {
                    var t = parameters.position
                    if(t.parent is PsiFile) {
                        t = t.prevSibling
                        var canBeTypeDeclaration = true
                        var canBeType = true
                        while(t != null) {
                            if(canBeTypeDeclaration && (t.elementType == Types.TYPEMODIFIER || t.elementType == Types.DECLARATIONS)) {
                                addTypeDeclarationCompletions(parameters, context, resultSet)
                                return
                            }
                            if(canBeTypeDeclaration && (t.elementType == Types.SEMI || t.elementType == Types.MODIFIERCLOSE || t.elementType == Types.MODIFIEROPEN)) {
                                canBeTypeDeclaration = false;
                            }
                            if(canBeType) {
                                if(t.elementType == Types.SEMI || t.elementType == Types.BRACKETOPEN) {
                                    addTypesComplition(parameters, context, resultSet)
                                    return
                                } else if(t.elementType != TokenType.WHITE_SPACE)
                                    canBeType = false
                            }
                            t = t.prevSibling
                        }
                    } else if(t.parent != null && t.parent.elementType == Types.TYPE) {
                        addTypesComplition(parameters, context, resultSet)
                        return
                    }
                }
            }
        )
    }

    fun addTypeDeclarationCompletions(@NotNull parameters: CompletionParameters,
                                      @NotNull context: ProcessingContext,
                                      @NotNull resultSet: CompletionResultSet) {
        if(parameters.position.parent != null && parameters.position.parent.elementType == Types.ANNOTATION_VALUE) {
            addTypeAnnotationParamsCompletions(parameters, context, resultSet)
        } else {
            resultSet.addElement(LookupElementBuilder.create("type"))
            resultSet.addElement(LookupElementBuilder.create("dictionary"))
            resultSet.addElement(LookupElementBuilder.create("abstract"))
            resultSet.addElement(LookupElementBuilder.create("config"))
            resultSet.addElement(LookupElementBuilder.create("transient"))
            resultSet.addElement(LookupElementBuilder.create("embeddable"))
            resultSet.addElement(LookupElementBuilder.create("singleton"))
            resultSet.addElement(LookupElementBuilder.create("extern"))
            resultSet.addElement(LookupElementBuilder.create("extendable"))
            var prefix =
                if (parameters.offset > 0 && parameters.editor.document.charsSequence[parameters.offset - 1] == '$')
                    ""
                else
                    "$"
            fillTypeAnnotations(resultSet, prefix)
            prefix =
                if (parameters.offset > 0 && parameters.editor.document.charsSequence[parameters.offset - 1] == '@') {
                    ""
                } else {
                    "@"
                }
            fillTypeHooks(resultSet, prefix)
        }
    }

    fun addTypesComplition(@NotNull parameters: CompletionParameters,
                           @NotNull context: ProcessingContext,
                           @NotNull resultSet: CompletionResultSet) {
        resultSet.addElement(LookupElementBuilder.create("int"))
        resultSet.addElement(LookupElementBuilder.create("float"))
        resultSet.addElement(LookupElementBuilder.create("bool"))
        resultSet.addElement(LookupElementBuilder.create("string"))
        resultSet.addElement(LookupElementBuilder.create("date"))
        resultSet.addElement(LookupElementBuilder.create("map[]"))
        resultSet.addElement(LookupElementBuilder.create("[]"))
        val p = parameters.editor.project
        val sdf = parameters.originalFile as SDFPsiFile
        var pckg = ""
        if(sdf != null)
            pckg = sdf.getPackage()
        if(p != null) {
            val types = Util.findTypes(p)
            for(t in types) {
                if(t.pckg == pckg)
                    resultSet.addElement(LookupElementBuilder.create(t.name))
                else
                    resultSet.addElement(LookupElementBuilder.create("${t.pckg}.${t.name}"))
            }
        }
    }
    override fun beforeCompletion(context: CompletionInitializationContext) {
//        context.dummyIdentifier = "\u001f";
        context.dummyIdentifier = "DuM_Id";
        super.beforeCompletion(context)
    }

    override fun handleAutoCompletionPossibility(context: AutoCompletionContext): AutoCompletionDecision? {
        return super.handleAutoCompletionPossibility(context)
    }

    fun addTypeAnnotationParamsCompletions(
        @NotNull parameters: CompletionParameters,
        @NotNull context: ProcessingContext,
        @NotNull resultSet: CompletionResultSet) {
        var annParam: String? = null
        var annotationTag : String? = null
        var ps = parameters.position
        while(ps != null) {
            if(ps.elementType == Types.ANN_PARAM) {
                if(ps.firstChild != null && ps.firstChild.elementType == Types.IDENTIFIER)
                    annParam = ps.firstChild.text
            }
            if(ps.elementType == Types.ANNOTATION_VALUES) {
                ps = ps.parent
                if(ps != null && ps.elementType == Types.ANNOTATION) {
                    ps = ps.firstChild
                    if(ps != null && ps.elementType == Types.ANNOTATIONTAG)
                        annotationTag = ps.text
                }
                break
            }
            if(ps.elementType == Types.TYPE_MODIFIER) {
                break
            }
            ps = ps.parent
        }
        if(annotationTag != null) {
            if (annParam != null)
                fillValuesForTagParam(resultSet, annotationTag, annParam)
            else
                fillParamsForTag(resultSet, annotationTag)
        }
    }
    fun fillTypeAnnotations(resultSet: CompletionResultSet, prefix: String) {
        resultSet.addElement(LookupElementBuilder.create("${prefix}vue"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}gql"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}db"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}mongo"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}find"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}go"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}js"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}gotags"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}config"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}deletable"))
    }

    fun fillTypeHooks(resultSet: CompletionResultSet, prefix: String) {
        resultSet.addElement(LookupElementBuilder.create("${prefix}create"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}change"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}start"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}time=\"@every 1m -> fnname\""))
    }

    fun fillAttrAnnotations(resultSet: CompletionResultSet, prefix: String) {
        resultSet.addElement(LookupElementBuilder.create("${prefix}vue"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}gql"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}historic"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}db"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}mongo"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}js"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}config"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}find"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}qualifier"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}qualified-by"))
    }

    fun fillAttrHooks(resultSet: CompletionResultSet, prefix: String) {
        resultSet.addElement(LookupElementBuilder.create("${prefix}set"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}resolve"))
    }
    fun fillParamsForTag(resultSet: CompletionResultSet, tag: String) {
        when(tag) {
            "\$go" -> {
                resultSet.addElement(LookupElementBuilder.create("name"))
                resultSet.addElement(LookupElementBuilder.create("attr-name"))
            }
            "\$ghl" -> {
                resultSet.addElement(LookupElementBuilder.create("name"))
                resultSet.addElement(LookupElementBuilder.create("skip"))
            }
            "\$historic" -> {
                resultSet.addElement(LookupElementBuilder.create("fields"))
            }
            "\$vue" -> {
                resultSet.addElement(LookupElementBuilder.create("ignore"))
                resultSet.addElement(LookupElementBuilder.create("type"))
            }
            "\$find" -> {
                resultSet.addElement(LookupElementBuilder.create("field"))
                resultSet.addElement(LookupElementBuilder.create("type"))
            }
            "\$call" -> {
                resultSet.addElement(LookupElementBuilder.create("name"))
                resultSet.addElement(LookupElementBuilder.create("js"))
            }
            "\$config" -> {
                resultSet.addElement(LookupElementBuilder.create("value"))
                resultSet.addElement(LookupElementBuilder.create("group"))
                resultSet.addElement(LookupElementBuilder.create("mutable"))
            }
            "\$deletable" -> {
                resultSet.addElement(LookupElementBuilder.create("field"))
                resultSet.addElement(LookupElementBuilder.create("ignore"))
            }
            "\$js" -> {
                resultSet.addElement(LookupElementBuilder.create("name"))
                resultSet.addElement(LookupElementBuilder.create("skip"))
                resultSet.addElement(LookupElementBuilder.create("title"))
                resultSet.addElement(LookupElementBuilder.create("icon"))
                resultSet.addElement(LookupElementBuilder.create("force"))
                resultSet.addElement(LookupElementBuilder.create("color"))
            }
        }
    }
    fun fillValuesForTagParam(resultSet: CompletionResultSet, tag: String, param: String) {
        when(tag) {
            "\$vue" -> {
                when(param) {
                    "type" -> {
                        resultSet.addElement(LookupElementBuilder.create("isodate"))
                        resultSet.addElement(LookupElementBuilder.create("date"))
                        resultSet.addElement(LookupElementBuilder.create("custom"))
                        resultSet.addElement(LookupElementBuilder.create("value"))
                    }
                }

            }
            "\$historic" -> {
                when (param) {
                    "fields" -> {
                        resultSet.addElement(LookupElementBuilder.create("timestamp"))
                        resultSet.addElement(LookupElementBuilder.create("uid"))
                        resultSet.addElement(LookupElementBuilder.create("uname"))
                        resultSet.addElement(LookupElementBuilder.create("source"))
                    }
                }
            }
            "\$find" -> {
                when(param) {
                    "field" -> {
                        resultSet.addElement(LookupElementBuilder.create("_deleted_"))
                    }
                    "type"-> {
                        resultSet.addElement(LookupElementBuilder.create("eq"))
                        resultSet.addElement(LookupElementBuilder.create("ne"))
                        resultSet.addElement(LookupElementBuilder.create("gt"))
                        resultSet.addElement(LookupElementBuilder.create("gte"))
                        resultSet.addElement(LookupElementBuilder.create("lt"))
                        resultSet.addElement(LookupElementBuilder.create("lte"))
                        resultSet.addElement(LookupElementBuilder.create("starts-with"))
                        resultSet.addElement(LookupElementBuilder.create("contains"))
                        resultSet.addElement(LookupElementBuilder.create("ignore"))
                        resultSet.addElement(LookupElementBuilder.create("is-null"))
                    }
                }
            }
            "\$call" -> {
                when(param) {
                    "js" -> {
                        fillBoolVariants(resultSet)
                    }
                }
            }
        }
    }
    fun fillMethodAnnotations(resultSet: CompletionResultSet, prefix: String) {
        resultSet.addElement(LookupElementBuilder.create("${prefix}call"))
        resultSet.addElement(LookupElementBuilder.create("${prefix}gql"))
    }
    fun fillBoolVariants(resultSet: CompletionResultSet) {
        resultSet.addElement(LookupElementBuilder.create("true"))
        resultSet.addElement(LookupElementBuilder.create("false"))
    }
}