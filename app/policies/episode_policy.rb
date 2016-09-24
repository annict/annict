# frozen_string_literal: true

class EpisodePolicy < ApplicationPolicy
  def hide?
    user.role.admin?
  end

  def destroy?
    user.role.admin?
  end
end
