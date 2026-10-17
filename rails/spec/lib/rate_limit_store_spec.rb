# typed: false
# frozen_string_literal: true

RSpec.describe RateLimitStore do
  include ActiveSupport::Testing::TimeHelpers

  let(:store) { described_class.new(period: 60) }
  # 2026-10-14 12:34:00 UTC から始まる1分の窓の通し番号
  let(:window_index) { Time.utc(2026, 10, 14, 12, 34).to_i / 60 }
  let(:name) { "rack::attack:#{window_index}:api:ip:192.0.2.1" }

  def stored_rows
    ActiveRecord::Base.connection.select_rows(<<~SQL.squish)
      SELECT key, window_start, count FROM rate_limits WHERE key LIKE 'api:ip:192.0.2.%' ORDER BY key, window_start
    SQL
  end

  describe "#increment" do
    it "窓の最初の呼び出しで行を作り、以降は同じ行の件数を増やすこと" do
      expect(store.increment(name, 1)).to eq(1)
      expect(store.increment(name, 1)).to eq(2)
      expect(store.increment(name, 3)).to eq(5)

      expect(stored_rows).to eq([["api:ip:192.0.2.1", Time.utc(2026, 10, 14, 12, 34), 5]])
    end

    it "窓の通し番号が異なるキーは別の行で数えること" do
      store.increment(name, 1)

      expect(store.increment("rack::attack:#{window_index + 1}:api:ip:192.0.2.1", 1)).to eq(1)
      expect(stored_rows.pluck(1)).to eq([Time.utc(2026, 10, 14, 12, 34), Time.utc(2026, 10, 14, 12, 35)])
    end

    it "IPv6アドレスのようにコロンを含む識別子もそのままキーに残すこと" do
      store.increment("rack::attack:#{window_index}:api:ip:192.0.2.2:2001:db8::1", 1)

      expect(stored_rows.first[0]).to eq("api:ip:192.0.2.2:2001:db8::1")
    end

    it "カウンターの読み書きに失敗したときはnilを返すこと" do
      allow(ActiveRecord::Base).to receive(:with_connection).and_raise(ActiveRecord::ConnectionNotEstablished)

      expect(store.increment(name, 1)).to be_nil
    end
  end

  describe "#read" do
    it "窓の件数を返し、行が無ければnilを返すこと" do
      expect(store.read(name)).to be_nil

      store.increment(name, 1)
      store.increment(name, 1)

      expect(store.read(name)).to eq(2)
    end
  end

  describe "#write" do
    it "行が無ければ窓を作り、既存の件数は上書きしないこと" do
      store.write(name, 7)
      expect(store.read(name)).to eq(7)

      store.write(name, 1)
      expect(store.read(name)).to eq(7)
    end

    it "増分の失敗後にrack-attackが書き込みを試みても既存の件数を保つこと" do
      store.increment(name, 7)
      allow(store).to receive(:increment).and_return(nil)
      cache = Rack::Attack::Cache.new
      cache.store = store

      travel_to(Time.utc(2026, 10, 14, 12, 34, 15)) do
        expect(cache.count("api:ip:192.0.2.1", 60)).to eq(1)
      end

      expect(store.read(name)).to eq(7)
    end
  end

  describe "#delete" do
    it "窓の行を削除すること" do
      store.increment(name, 1)

      store.delete(name)

      expect(store.read(name)).to be_nil
    end
  end

  it "窓の通し番号を含まないキーはArgumentErrorにすること" do
    expect { store.read("rack::attack:allow2ban:ban:192.0.2.1") }.to raise_error(ArgumentError)
  end
end
