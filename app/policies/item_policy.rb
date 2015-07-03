class ItemPolicy < ApplicationPolicy
  def update?
    user.role.editor? || user.role.admin?
  end

  def destroy?
    user.role.admin?
  end
end
