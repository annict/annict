module Annict
  module Exceptions
    # エピソードの一括記録で episodes.id 以外の値でEpisodeを検索したときに発生するエラー
    class UnexpectedEpisodeIdsError < StandardError; end
  end
end
