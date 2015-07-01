class WorkPolicy < ApplicationPolicy
  def update?
    user.role.admin?
  end

  def destroy?
    user.role.admin?
  end
end
