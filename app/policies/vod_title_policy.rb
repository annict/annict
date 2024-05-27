# typed: false
# frozen_string_literal: true

class VodTitlePolicy < ApplicationPolicy
  def unpublish?
    user.present? && user.admin?
  end
end
