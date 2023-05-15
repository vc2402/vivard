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

public class SDFTypeImpl extends ASTWrapperPsiElement implements SDFType {

  public SDFTypeImpl(@NotNull ASTNode node) {
    super(node);
  }

  public void accept(@NotNull SDFVisitor visitor) {
    visitor.visitType(this);
  }

  @Override
  public void accept(@NotNull PsiElementVisitor visitor) {
    if (visitor instanceof SDFVisitor) accept((SDFVisitor)visitor);
    else super.accept(visitor);
  }

  @Override
  @Nullable
  public SDFArrayType getArrayType() {
    return findChildByClass(SDFArrayType.class);
  }

  @Override
  @Nullable
  public SDFMapType getMapType() {
    return findChildByClass(SDFMapType.class);
  }

  @Override
  @Nullable
  public SDFSimpleType getSimpleType() {
    return findChildByClass(SDFSimpleType.class);
  }

}
