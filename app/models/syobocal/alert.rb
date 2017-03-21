# == Schema Information
#
# Table name: syobocal_alerts
#
#  id              :integer          not null, primary key
#  work_id         :integer
#  kind            :integer          not null
#  sc_prog_item_id :integer
#  sc_sub_title    :string(255)
#  sc_prog_comment :string(255)
#  created_at      :datetime
#  updated_at      :datetime
#
# Indexes
#
#  index_syobocal_alerts_on_kind             (kind)
#  index_syobocal_alerts_on_sc_prog_item_id  (sc_prog_item_id)
#

class Syobocal::Alert < ApplicationRecord
  extend Enumerize

  enumerize :kind, in: { special_program: 0 }, scope: true

  belongs_to :work


  def self.new_special_program?(sc_prog_item_id)
    alert = with_kind(:special_program).find_by(sc_prog_item_id: sc_prog_item_id)
    alert.blank?
  end
end
