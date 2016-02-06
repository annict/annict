class EditRequestCommentPolicy < ApplicationPolicy
  def create?
    signed_in?
  end

  def update?
    signed_in? && user == record.user
  end

  def destroy?
    signed_in? && user == record.user
  end
end
