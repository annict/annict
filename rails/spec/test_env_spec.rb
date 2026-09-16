# typed: false
# frozen_string_literal: true

RSpec.describe TestEnv do
  describe ".apply!" do
    it "外部から季節やCookieのドメインが渡されていても、.env.testの値で上書きすること" do
      env = {
        "ANNICT_CURRENT_SEASON" => "2025-autumn",
        "ANNICT_COOKIE_DOMAIN" => ".example.com",
        "ANNICT_URL" => "https://example.com"
      }

      TestEnv.apply!(env:)

      expect(env["ANNICT_CURRENT_SEASON"]).to eq("2017-winter")
      expect(env["ANNICT_COOKIE_DOMAIN"]).to eq("")
      expect(env["ANNICT_URL"]).to eq("http://localhost:4000")
    end

    it "固定対象でないDB接続情報は、外部から渡された値を引き継ぐこと" do
      env = {"ANNICT_POSTGRES_HOST" => "localhost"}

      TestEnv.apply!(env:)

      expect(env["ANNICT_POSTGRES_HOST"]).to eq("localhost")
    end

    it "実行中のプロセスには.env.testの固定値が適用されていること" do
      expected = Dotenv.parse(TestEnv::ENV_FILE).slice(*TestEnv::KEYS)

      expect(ENV.to_h.slice(*TestEnv::KEYS)).to eq(expected)
    end

    it "固定対象のキーが.env.testに無い場合、エラーにすること" do
      Tempfile.create(["env", ".test"]) do |file|
        file.write("ANNICT_URL=http://localhost:4000\n")
        file.flush

        expect { TestEnv.apply!(env: {}, path: file.path) }.to raise_error(KeyError, /ANNICT_CURRENT_SEASON/)
      end
    end
  end
end
