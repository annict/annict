# == Schema Information
#
# Table name: cover_images
#
#  id         :integer          not null, primary key
#  work_id    :integer          not null
#  file_name  :string(510)      not null
#  location   :string(510)      not null
#  created_at :datetime
#  updated_at :datetime
#
# Indexes
#
#  cover_images_work_id_idx  (work_id)
#

class CoverImage < ActiveRecord::Base
  belongs_to :work
end
