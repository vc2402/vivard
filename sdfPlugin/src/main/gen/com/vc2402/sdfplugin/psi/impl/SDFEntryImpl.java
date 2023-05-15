// This is a generated file. Not intended for manual editing.
package com.vc2402.sdfplugin.psi.impl;

import java.util.List;
import org.jetbrains.annotations.*;
import com.intellij.lang.ASTNode;
import com.intellij.psi.PsiElement;
import com.intellij.psi.PsiElementVisitor;
import com.intellij.psi.util.PsiTreeUtil;
import static com.vc2402.sdfplugin.psi.Types.*;
import com.intellij.extapi.psi.ASTWrapperPsiElement;
import com.vc2402.sdfplugin.psi.*;

public class SDFEntryImpl extends ASTWrapperPsiElement implements SDFEntry {

  public SDFEntryImpl(@NotNull ASTNode node) {
    super(node);
  }

  public void accept(@NotNull SDFVisitor visitor) {
    visitor.visitEntry(this);
  }

  @Override
  public void accept(@NotNull PsiElementVisitor visitor) {
    if (visitor instanceof SDFVisitor) accept((SDFVisitor)visitor);
    else super.accept(visitor);
  }

  @Override
  @Nullable
  public SDFAttrModifiers getAttrModifiers() {
    return findChildByClass(SDFAttrModifiers.class);
  }

  @Override
  @Nullable
  public SDFField getField() {
    return findChildByClass(SDFField.class);
  }

  @Override
  @Nullable
  public SDFMethod getMethod() {
    return findChildByClass(SDFMethod.class);
  }

}
