# typed: false
# frozen_string_literal: true

class WorkImage < ApplicationRecord
  T.unsafe(self).include WorkImageUploader::Attachment.new(:image)
  include ImageUploadable
  include AnimeIdCopyable

  self.ignored_columns = %w[color_rgb]

  validates :copyright, presence: true

  belongs_to :work, touch: true
  belongs_to :user

  copy_anime_id_from :work
end
