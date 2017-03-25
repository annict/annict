# frozen_string_literal: true

namespace :tmp do
  task update_channels: :environment do
    [
      "Amazonビデオ",
      "Crunchyroll",
      "DAISUKI",
      "Funimation",
      "Hulu",
      "Netflix",
      "U-NEXT"
    ].each do |name|
      puts name
      Channel.where(name: name).first_or_create!(streaming_service: true)
    end

    sc_chids = [234, 107, 165]
    Channel.where(sc_chid: sc_chids).update_all(streaming_service: true)

    Channel.find_by(sc_chid: 132).update(aasm_state: :hidden)
  end
end
