# typed: false
# frozen_string_literal: true

class ChannelPolicy < ApplicationPolicy
  def create?
    user.present? && user.admin?
  end

  def update?
    user.present? && user.admin?
  end

  def destroy?
    user.present? && user.admin?
  end

  def unpublish?
    user.present? && user.admin?
  end
end
