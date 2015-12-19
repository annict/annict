class PersonPolicy < ApplicationPolicy
  def create?
    user.present? && user.role.editor? || user.role.admin?
  end

  def update?
    user.present? && user.role.editor? || user.role.admin?
  end

  def hide?
    user.present? && user.role.admin?
  end

  def destroy?
    user.present? && user.role.admin?
  end
end
