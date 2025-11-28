# typed: false
# frozen_string_literal: true

class UserlandProjectPolicy < ApplicationPolicy
  def update?
    user.userland_project_member?(record)
  end

  def destroy?
    user.userland_project_member?(record)
  end
end
