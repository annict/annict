module DraftCommon
  extend ActiveSupport::Concern

  included do
    has_one :edit_request, as: :draft_resource

    accepts_nested_attributes_for :edit_request

    def diffs
      origin_hash = edit_request.draft_resource.
        try!(:origin).
        try!(:to_diffable_hash).presence || {}
      draft_hash = edit_request.draft_resource.to_diffable_hash

      HashDiff.diff(origin_hash, draft_hash)
    end

    def origin_values
      edit_request.draft_resource.
        try!(:origin).
        try!(:decorate).
        try!(:to_values).
        presence || {}
    end

    def draft_values
      edit_request.draft_resource.decorate.to_values
    end
  end
end
