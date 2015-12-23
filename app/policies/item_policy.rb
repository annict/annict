class ItemPolicy < ApplicationPolicy
  def create?
    user.committer?
  end

  def update?
    user.committer?
  end

  def destroy?
    user.role.admin?
  end
end
