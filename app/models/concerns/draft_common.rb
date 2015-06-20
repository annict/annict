module DraftCommon
  extend ActiveSupport::Concern

  included do
    has_one :edit_request, as: :draft_resource

    accepts_nested_attributes_for :edit_request

    after_save :update_diffs_and_values

    private

    def update_diffs_and_values
      origin_hash = edit_request.draft_resource.
                      try(:origin).
                      try(:to_diffable_hash).presence || {}
      draft_hash = edit_request.draft_resource.to_diffable_hash

      diffs = HashDiff.diff(origin_hash, draft_hash).delete_if { |diff| diff[2].blank? }
      origin_values = edit_request.draft_resource.
                        try(:origin).try(:decorate).try(:to_values)
      draft_values = edit_request.draft_resource.decorate.to_values

      edit_request.update(
        diffs: diffs,
        draft_values: draft_values,
        origin_values: origin_values
      )
    end
  end
end
