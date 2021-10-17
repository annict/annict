# frozen_string_literal: true

describe Forms::EpisodeRecordForm do
  let(:user) { create :registered_user }
  let(:episode) { create :episode }

  context "バリデーションエラーになったとき" do
    it "エラー内容を返すこと" do
      form = Forms::EpisodeRecordForm.new(user: user, episode: episode)
      form.attributes = {
        comment: "a" * (1_048_596 + 1), # 文字数制限 (1,048,596文字) 以上の感想を書く
        rating: "good",
        share_to_twitter: false
      }

      expect(form.valid?).to eq false
      expect(form.errors.count).to eq 1
      expect(form.errors.full_messages.first).to eq "感想は1048596文字以内で入力してください"
    end
  end
end
