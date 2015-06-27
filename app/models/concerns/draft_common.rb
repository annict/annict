module DraftCommon
  extend ActiveSupport::Concern

  included do
    has_one :edit_request, as: :draft_resource, dependent: :destroy

    accepts_nested_attributes_for :edit_request

    def diffs
      origin_hash = edit_request.draft_resource.
                      try(:origin).
                      try(:to_diffable_hash).presence || {}
      draft_hash = edit_request.draft_resource.to_diffable_hash

      HashDiff.diff(origin_hash, draft_hash).delete_if { |diff| diff[2].blank? }
    end

    def origin_values
      edit_request.draft_resource.try(:origin).try(:decorate).try(:to_values)
    end

    def draft_values
      edit_request.draft_resource.decorate.to_values
    end
  end
end
