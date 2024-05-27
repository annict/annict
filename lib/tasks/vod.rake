# typed: false
# frozen_string_literal: true

namespace :vod do
  task import: :environment do
    [
      Annict::VodImporter::BandaiChannel,
      Annict::VodImporter::NicoNicoChannel,
      Annict::VodImporter::DAnimeStore,
      Annict::VodImporter::AmazonVideo
    ].each(&:import)
  end
end
