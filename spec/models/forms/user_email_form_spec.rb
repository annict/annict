# frozen_string_literal: true

describe Forms::UserEmailForm do
  let(:email) { "example@example.com" }
  let(:form) do
    Forms::UserEmailForm.new(
      email: email
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
    end
  end
end
