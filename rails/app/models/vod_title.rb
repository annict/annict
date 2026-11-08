# typed: false
# frozen_string_literal: true

class VodTitle < ApplicationRecord
  include SoftDeletable
  include AnimeIdCopyable

  belongs_to :channel
  belongs_to :work, optional: true

  copy_anime_id_from :work

  def import_csv
    [
      channel_id,
      nil,
      nil,
      code,
      name
    ].join(",")
  end
end
