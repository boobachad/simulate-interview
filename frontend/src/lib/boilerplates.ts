export const CPP_BOILERPLATE = `#include <bits/stdc++.h>
using namespace std;

void solve() {
    // ============================================
    // START: Write your solution code here
    // ============================================
    
    int a, b;
    cin >> a >> b;
    cout << a + b << endl;
    
    // ============================================
    // END: Write your solution code above
    // ============================================
}

// DO NOT MODIFY BELOW THIS LINE
int main() {
    ios_base::sync_with_stdio(false);
    cin.tie(NULL);
    int t;
    cin >> t;
    while(t--) solve();
    return 0;
}`;

export const PYTHON_BOILERPLATE = `def solve():
    # ============================================
    # START: Write your solution code here
    # ============================================
    
    a, b = map(int, input().split())
    print(a + b)
    
    # ============================================
    # END: Write your solution code above
    # ============================================

# DO NOT MODIFY BELOW THIS LINE
if __name__ == "__main__":
    t = int(input())
    for _ in range(t):
        solve()`;

export const JAVA_BOILERPLATE = `import java.util.*;
import java.io.*;

public class Main {
    public static void solve(Scanner sc) {
        // ============================================
        // START: Write your solution code here
        // ============================================
        
        int a = sc.nextInt();
        int b = sc.nextInt();
        System.out.println(a + b);
        
        // ============================================
        // END: Write your solution code above
        // ============================================
    }
    
    // DO NOT MODIFY BELOW THIS LINE
    public static void main(String[] args) {
        Scanner sc = new Scanner(System.in);
        int t = sc.nextInt();
        while (t-- > 0) {
            solve(sc);
        }
        sc.close();
    }
}`;

export const JAVASCRIPT_BOILERPLATE = `const readline = require('readline');
const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
});

let lines = [];
rl.on('line', (line) => {
    lines.push(line);
}).on('close', () => {
    const t = parseInt(lines[0]);
    let idx = 1;
    
    for (let i = 0; i < t; i++) {
        // ============================================
        // START: Write your solution code here
        // ============================================
        
        const [a, b] = lines[idx++].split(' ').map(Number);
        console.log(a + b);
        
        // ============================================
        // END: Write your solution code above
        // ============================================
    }
});`;

export const BOILERPLATES: Record<string, string> = {
  cpp: CPP_BOILERPLATE,
  python: PYTHON_BOILERPLATE,
  java: JAVA_BOILERPLATE,
  javascript: JAVASCRIPT_BOILERPLATE,
};
