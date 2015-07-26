class EditRequestCommentPolicy < ApplicationPolicy
  def create?
    user.present?
  end
end
