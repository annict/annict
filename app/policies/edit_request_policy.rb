class EditRequestPolicy < ApplicationPolicy
  def publish?
    user.role.editor? || user.role.admin?
  end

  def close?
    user.role.editor? || user.role.admin? || user == record.user
  end
end
