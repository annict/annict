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
#  main_color_hex          :string           default("ffffff"), not null
#
# Indexes
#
#  index_work_images_on_user_id  (user_id)
#  index_work_images_on_work_id  (work_id)
#

class WorkImage < ApplicationRecord
  has_attached_file :attachment

  validates :attachment,
    attachment_presence: true,
    attachment_content_type: { content_type: /\Aimage/ }
  validates :asin, asin: true
  validates_with AsinOrCopyrightValidator

  belongs_to :work
  belongs_to :user

  after_save :set_color_rgb

  def attachment_relative_path(convert_type = "master")
    if Rails.env.production?
      "/#{attachment.path(convert_type)}"
    else
      attachment.url(convert_type)
    end
  end

  def attachment_absolute_path(convert_type = "master")
    if Rails.env.production?
      attachment_url(convert_type)
    else
      attachment.path(convert_type)
    end
  end

  def attachment_url(convert_type = "master")
    "#{ENV.fetch('ANNICT_FILE_STORAGE_URL')}#{attachment_relative_path(convert_type)}"
  end

  def colors
    return @colors if @colors.present?

    colors = Miro::DominantColors.new(attachment_url)
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
    colors = Miro::DominantColors.new(attachment_absolute_path)
    color_rgb = colors.to_rgb.map { |c| c.join(",") }.first
    update_column(:color_rgb, color_rgb)
  end
end
