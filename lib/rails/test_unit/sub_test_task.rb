# frozen_string_literal: true

# DraperをRails 5で動かすために必要なモンキーパッチ
# https://github.com/drapergem/draper/issues/681
module Rails
  class SubTestTask < Rake::TestTask; end
end
