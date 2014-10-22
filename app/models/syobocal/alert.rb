class Syobocal::Alert < ActiveRecord::Base
  extend Enumerize

  enumerize :kind, in: { special_program: 0 }, scope: true

  belongs_to :work


  def self.new_special_program?(sc_prog_item_id)
    alert = with_kind(:special_program).find_by(sc_prog_item_id: sc_prog_item_id)
    alert.blank?
  end
end
