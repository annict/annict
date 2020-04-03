# frozen_string_literal: true

class DbResourcePublishingPolicy < ApplicationPolicy
  def create?
    user.present? && user.committer?
  end

  def destroy?
    create?
  end
end
