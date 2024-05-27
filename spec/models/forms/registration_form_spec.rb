# typed: false
# frozen_string_literal: true

describe Forms::RegistrationForm do
  let(:email) { "example@example.com" }
  let(:token) { "foobar" }
  let(:username) { "example" }
  let(:terms_and_privacy_policy_agreement) { true }
  let(:form) do
    Forms::RegistrationForm.new(
      email: email,
      token: token,
      username: username,
      terms_and_privacy_policy_agreement: terms_and_privacy_policy_agreement
    )
  end

  describe "#valid?" do
    context "when attributes are valid" do
      it do
        expect(form.valid?).to be true
        expect(form.errors.full_messages).to eq []
      end
    end

    context "when attributes are invalid" do
      context "email" do
        context "exists" do
          let(:user) { create :registered_user }
          let(:email) { user.email }

          it do
            expect(form.valid?).to be false
            expect(form.errors.full_messages).to eq ["メールアドレスはすでに存在します"]
          end
        end

        context "exists as different character case" do
          let(:user) { create :registered_user }
          let(:email) { user.email.upcase }

          it do
            expect(form.valid?).to be false
            expect(form.errors.full_messages).to eq ["メールアドレスはすでに存在します"]
          end
        end

        context "is empty" do
          let(:email) { "" }

          it do
            expect(form.valid?).to be false
            expect(form.errors.full_messages).to eq %w[メールアドレスを入力してください メールアドレスは不正な値です]
          end
        end

        context "is invalid format" do
          let(:email) { "not-email" }

          it do
            expect(form.valid?).to be false
            expect(form.errors.full_messages).to eq ["メールアドレスは不正な値です"]
          end
        end
      end

      context "username" do
        context "exists" do
          let(:user) { create :registered_user }
          let(:username) { user.username }

          it do
            expect(form.valid?).to be false
            expect(form.errors.full_messages).to eq ["ユーザ名はすでに存在します"]
          end
        end

        context "exists as different character case" do
          let(:user) { create :registered_user }
          let(:username) { user.username.upcase }

          it do
            expect(form.valid?).to be false
            expect(form.errors.full_messages).to eq ["ユーザ名はすでに存在します"]
          end
        end

        context "is empty" do
          let(:username) { "" }

          it do
            expect(form.valid?).to be false
            expect(form.errors.full_messages).to eq %w[ユーザ名を入力してください ユーザ名は不正な値です]
          end
        end

        context "is invalid format" do
          let(:username) { "invalid-username" }

          it do
            expect(form.valid?).to be false
            expect(form.errors.full_messages).to eq ["ユーザ名は不正な値です"]
          end
        end

        context "is too long" do
          let(:username) { "usernameusernameusernameusernameusernameusernameusernameusername" }

          it do
            expect(form.valid?).to be false
            expect(form.errors.full_messages).to eq ["ユーザ名は20文字以内で入力してください"]
          end
        end
      end

      context "terms_and_privacy_policy_agreement" do
        context "is empty" do
          let(:terms_and_privacy_policy_agreement) { nil }

          it do
            expect(form.valid?).to be false
            expect(form.errors.full_messages).to eq ["利用規約とプライバシーポリシーを入力してください"]
          end
        end

        context "is false" do
          let(:terms_and_privacy_policy_agreement) { false }

          it do
            expect(form.valid?).to be false
            expect(form.errors.full_messages).to eq %w[
              利用規約とプライバシーポリシーを入力してください
              利用規約とプライバシーポリシーを受諾してください
            ]
          end
        end
      end
    end
  end
end
