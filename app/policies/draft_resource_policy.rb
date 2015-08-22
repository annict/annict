class DraftResourcePolicy < ApplicationPolicy
  def update?
    user.present? &&
    user == record.edit_request.user &&
    !record.edit_request.published? &&
    !record.edit_request.closed?
  end
end
