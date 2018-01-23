# frozen_string_literal: true

class VodTitlePolicy < ApplicationPolicy
  def hide?
    user.present? && user.role.admin?
  end
end
