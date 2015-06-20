class DraftResourcePolicy < ApplicationPolicy
  def update?
    !record.edit_request.published? && !record.edit_request.closed?
  end
end
