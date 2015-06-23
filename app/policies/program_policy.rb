class ProgramPolicy < ApplicationPolicy
  def destroy?
    user.role.admin?
  end
end
