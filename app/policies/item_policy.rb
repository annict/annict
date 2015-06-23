class ItemPolicy < ApplicationPolicy
  def destroy?
    user.role.admin?
  end
end
