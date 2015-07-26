class DraftResourcePolicy < ApplicationPolicy
  def update?
    user.present? && !record.edit_request.published? && !record.edit_request.closed?
  end
end
