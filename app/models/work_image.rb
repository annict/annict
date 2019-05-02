# frozen_string_literal: true
# == Schema Information
#
# Table name: work_images
#
#  id                      :integer          not null, primary key
#  work_id                 :integer          not null
#  user_id                 :integer          not null
#  attachment_file_name    :string           not null
#  attachment_file_size    :integer          not null
#  attachment_content_type :string           not null
#  attachment_updated_at   :datetime         not null
#  copyright               :string           default(""), not null
#  asin                    :string           default(""), not null
#  created_at              :datetime         not null
#  updated_at              :datetime         not null
#  color_rgb               :string           default("255,255,255"), not null
#  image_data              :text
#
# Indexes
#
#  index_work_images_on_user_id  (user_id)
#  index_work_images_on_work_id  (work_id)
#

class WorkImage < ApplicationRecord
  include WorkImageUploader::Attachment.new(:image)

  validates :asin, asin: true
  validates_with AsinOrCopyrightValidator

  belongs_to :work
  belongs_to :user

  # Disable until the Paperclip -> Shrine migration is done
  # after_save :set_color_rgb

  def colors
    return @colors if @colors.present?

    colors = Miro::DominantColors.new(image_url(:master))
    @colors = colors.to_rgb.map { |c| c.join(",") }
    @colors
  end

  def text_color_rgb(light: "255,255,255", dark: "0,0,0")
    # https://stackoverflow.com/questions/3942878/how-to-decide-font-color-in-white-or-black-depending-on-background-color
    red, green, blue = color_rgb.split(",").map(&:to_i)
    (red * 0.299 + green * 0.587 + blue * 0.114) > 186 ? dark : light
  end

  private

  def set_color_rgb
    colors = Miro::DominantColors.new(image_url(:master))
    color_rgb = colors.to_rgb.map { |c| c.join(",") }.first
    update_column(:color_rgb, color_rgb)
  end
end
