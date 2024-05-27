# typed: false
# frozen_string_literal: true

namespace :sorbet do
  task update: :environment do
    system "bin/tapioca gem"
    system "bin/tapioca dsl"
    system "bin/tapioca todo"

    # Annictでは現状古いgraphql gemを使用しているため、`tapioca annotations` で取得した型ファイルに指定された引数と
    # 古いgraphql gemの引数が異なりエラーになる。そのため一旦コメントアウトする
    # https://github.com/Shopify/rbi-central/blob/eeb5b4f519d646eeae58dc6d8d9cccf83f7e2ef2/rbi/annotations/graphql.rbi#L5-L6
    # system "bin/tapioca annotations"
  end
end
