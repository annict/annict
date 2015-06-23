class WorkPolicy < ApplicationPolicy
  def destroy?
    user.role.admin?
  end
end
