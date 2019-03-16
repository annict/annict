# frozen_string_literal: true

class ChannelPolicy < ApplicationPolicy
  def create?
    user.role.admin?
  end

  def update?
    user.role.admin?
  end

  def hide?
    user.role.admin?
  end

  def destroy?
    user.role.admin?
  end
end
