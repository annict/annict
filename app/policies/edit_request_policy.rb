class EditRequestPolicy < ApplicationPolicy
  def publish?
    user.present? && (user.role.editor? || user.role.admin?)
  end

  def close?
    user.present? && (user.role.editor? || user.role.admin? || user == record.user)
  end
end
